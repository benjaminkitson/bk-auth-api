package main

import (
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigateway"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscertificatemanager"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscognito"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsroute53"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsroute53targets"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssecretsmanager"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsssm"
	awslambdago "github.com/aws/aws-cdk-go/awscdklambdagoalpha/v2"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type CdkWorkshopStackProps struct {
	awscdk.StackProps
}

func defaultAuthLambdaProps(path string) *awslambdago.GoFunctionProps {
	return &awslambdago.GoFunctionProps{
		Architecture: awslambda.Architecture_ARM_64(),
		Description:  jsii.String("Handler for Auth"),
		Tracing:      awslambda.Tracing_ACTIVE,
		Bundling: &awslambdago.BundlingOptions{
			GoBuildFlags: jsii.Strings(`-trimpath -buildvcs=false`),
		},
		Runtime:    awslambda.Runtime_PROVIDED_AL2(),
		MemorySize: jsii.Number(256),
		Timeout:    awscdk.Duration_Minutes(jsii.Number(5)),
		Entry:      &path,
	}
}

func NewCdkWorkshopStack(scope constructs.Construct, id string, props *CdkWorkshopStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	pool := awscognito.NewUserPool(stack, jsii.String("testPool"), &awscognito.UserPoolProps{
		UserPoolName:      jsii.String("Test User Pool"),
		SelfSignUpEnabled: jsii.Bool(true),
		SignInAliases: &awscognito.SignInAliases{
			Email: jsii.Bool(true),
		},
		AutoVerify: &awscognito.AutoVerifiedAttrs{
			Email: jsii.Bool(true),
		},
		PasswordPolicy: &awscognito.PasswordPolicy{
			MinLength:        jsii.Number(8),
			RequireLowercase: jsii.Bool(true),
			RequireUppercase: jsii.Bool(true),
			RequireDigits:    jsii.Bool(true),
			RequireSymbols:   jsii.Bool(false),
		},
		AccountRecovery: awscognito.AccountRecovery_EMAIL_ONLY,
		RemovalPolicy:   awscdk.RemovalPolicy_DESTROY,
		UserVerification: &awscognito.UserVerificationConfig{
			EmailSubject: jsii.String("You need to verify your email"),
			EmailBody:    jsii.String("Thanks for signing up! Your verification code is {####}"),
			EmailStyle:   awscognito.VerificationEmailStyle_CODE,
		},
	})

	pool.AddClient(jsii.String("test-client"), &awscognito.UserPoolClientOptions{
		UserPoolClientName: jsii.String("test-pool-client"),
		AuthFlows: &awscognito.AuthFlow{
			UserPassword: jsii.Bool(true),
		},
	})

	c := awssecretsmanager.NewSecret(stack, jsii.String("cognitoClientId"), &awssecretsmanager.SecretProps{
		//TODO: change this back to "COGNITO_CLIENT" once scheduled deletion has occured
		SecretName: jsii.String("COGNITO_CLIENT"),
	})

	// Fallback lambda
	fallbackLambda := awslambdago.NewGoFunction(stack, jsii.String("fallbackHandler"), defaultAuthLambdaProps("../lambda/fallback"))

	// Sign in
	signInLambda := awslambdago.NewGoFunction(stack, jsii.String("signInHandler"), defaultAuthLambdaProps("../lambda/signin"))
	c.GrantRead(signInLambda, nil)

	// Sign up
	signUpLambda := awslambdago.NewGoFunction(stack, jsii.String("signUpHandler"), defaultAuthLambdaProps("../lambda/signup"))
	c.GrantRead(signUpLambda, nil)

	// Verify Email
	userAPIParamName := jsii.String("/http-endpoints/user-api")

	verifyEmailLambda := awslambdago.NewGoFunction(stack, jsii.String("verifyEmailHandler"), defaultAuthLambdaProps("../lambda/verify"))
	verifyEmailLambda.AddEnvironment(jsii.String("USER_API_PARAMETER_NAME"), userAPIParamName, &awslambda.EnvironmentOptions{})
	verifyEmailLambda.AddToRolePolicy(awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Effect:  awsiam.Effect_ALLOW,
		Actions: jsii.Strings("execute-api:Invoke"),
		Resources: jsii.Strings(
			"arn:aws:execute-api:eu-west-2:905418429454:6blz968hz8/*/*/*",
		),
	}))
	c.GrantRead(verifyEmailLambda, nil)

	p := awsssm.NewStringParameter(stack, jsii.String("userAPIEndpoint"), &awsssm.StringParameterProps{
		ParameterName: userAPIParamName,
		// Added after the fact in the console
		StringValue: jsii.String("/"),
	})

	p.GrantRead(verifyEmailLambda)

	authApi := awsapigateway.NewLambdaRestApi(stack, jsii.String("Endpoint"), &awsapigateway.LambdaRestApiProps{
		DomainName: &awsapigateway.DomainNameOptions{
			// TODO: This needs putting somewhere else, defeats the purpose of SSM above
			DomainName: jsii.String("auth.benjaminkitson.com"),
			Certificate: awscertificatemanager.Certificate_FromCertificateArn(
				stack,
				jsii.String("benjaminkitson-certificate"),
				jsii.String("arn:aws:acm:eu-west-2:905418429454:certificate/42197bf4-d86d-404a-87a6-748c4858d916"),
			),
		},
		DisableExecuteApiEndpoint: jsii.Bool(true),
		RestApiName:               jsii.String("bk-auth"),
		Handler:                   fallbackLambda,
	})

	signUp := authApi.Root().AddResource(jsii.String("signup"), &awsapigateway.ResourceOptions{})
	signUp.AddMethod(jsii.String("POST"), awsapigateway.NewLambdaIntegration(signUpLambda, &awsapigateway.LambdaIntegrationOptions{}), &awsapigateway.MethodOptions{
		AuthorizationType: awsapigateway.AuthorizationType_IAM,
	})

	signIn := authApi.Root().AddResource(jsii.String("signin"), &awsapigateway.ResourceOptions{})
	signIn.AddMethod(jsii.String("POST"), awsapigateway.NewLambdaIntegration(signInLambda, &awsapigateway.LambdaIntegrationOptions{}), &awsapigateway.MethodOptions{
		AuthorizationType: awsapigateway.AuthorizationType_IAM,
	})

	verifyEmail := authApi.Root().AddResource(jsii.String("verify"), &awsapigateway.ResourceOptions{})
	verifyEmail.AddMethod(jsii.String("POST"), awsapigateway.NewLambdaIntegration(signInLambda, &awsapigateway.LambdaIntegrationOptions{}), &awsapigateway.MethodOptions{
		AuthorizationType: awsapigateway.AuthorizationType_IAM,
	})

	z := awsroute53.HostedZone_FromLookup(stack, jsii.String("zone"), &awsroute53.HostedZoneProviderProps{
		DomainName: jsii.String("benjaminkitson.com"),
	})

	awsroute53.NewARecord(stack, jsii.String("authRecord"), &awsroute53.ARecordProps{
		Zone:       z,
		RecordName: jsii.String("auth"),
		Target:     awsroute53.RecordTarget_FromAlias(awsroute53targets.NewApiGateway(authApi)),
	})

	// awscloudfront.NewDistribution(stack, jsii.String("myDist"), &awscloudfront.DistributionProps{
	// 	DefaultBehavior: &awscloudfront.BehaviorOptions{
	// 		Origin: awscloudfrontorigins.NewRestApiOrigin(pokedexApi, &awscloudfrontorigins.RestApiOriginProps{}),
	// 	},
	// })

	// userApi := awsapigateway.NewLambdaRestApi(stack, jsii.String("Endpoint"), &awsapigateway.LambdaRestApiProps{
	// 	DomainName: &awsapigateway.DomainNameOptions{
	// 		DomainName: jsii.String("api.benjaminkitson.com"),
	// 		Certificate: awscertificatemanager.Certificate_FromCertificateArn(
	// 			stack,
	// 			jsii.String("benjaminkitson-certificate"),
	// 			jsii.String("arn:aws:acm:eu-west-2:905418429454:certificate/42197bf4-d86d-404a-87a6-748c4858d916"),
	// 		),
	// 	},
	// 	DisableExecuteApiEndpoint: jsii.Bool(true),
	// 	RestApiName:               jsii.String("bk-api"),
	// 	Handler:                   fallbackLambda,
	// 	Proxy:                     jsii.Bool(false),
	// })

	// users := userApi.Root().AddResource(jsii.String("user"), &awsapigateway.ResourceOptions{})

	// proxy := users.AddProxy(&awsapigateway.ProxyResourceOptions{
	// 	DefaultIntegration: awsapigateway.NewLambdaIntegration(fallbackLambda, &awsapigateway.LambdaIntegrationOptions{}),
	// 	AnyMethod:          jsii.Bool(false),
	// })

	// proxy.AddMethod(jsii.String("POST"), awsapigateway.NewLambdaIntegration(userLambda, &awsapigateway.LambdaIntegrationOptions{}), &awsapigateway.MethodOptions{
	// 	AuthorizationType: awsapigateway.AuthorizationType_IAM,
	// })

	return stack
}

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	NewCdkWorkshopStack(app, "AuthTestStack", &CdkWorkshopStackProps{
		awscdk.StackProps{
			Env: env(),
		},
	})

	app.Synth(nil)
}

// env determines the AWS environment (account+region) in which our stack is to
// be deployed. For more information see: https://docs.aws.amazon.com/cdk/latest/guide/environments.html
func env() *awscdk.Environment {
	// If unspecified, this stack will be "environment-agnostic".
	// Account/Region-dependent features and context lookups will not work, but a
	// single synthesized template can be deployed anywhere.
	//---------------------------------------------------------------------------
	// return nil

	// Uncomment if you know exactly what account and region you want to deploy
	// the stack to. This is the recommendation for production stacks.
	//---------------------------------------------------------------------------
	// return &awscdk.Environment{
	//  Account: jsii.String("123456789012"),
	//  Region:  jsii.String("us-east-1"),
	// }

	// Uncomment to specialize this stack for the AWS Account and Region that are
	// implied by the current CLI configuration. This is recommended for dev
	// stacks.
	//---------------------------------------------------------------------------
	return &awscdk.Environment{
		Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
		Region:  jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
	}
}
