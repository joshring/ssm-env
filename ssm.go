package ssm

import (
	"context"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

// SetEnvVarsUsingSSM sets environment variables using AWS SSM.
//
// All parameters are fetched from SSM using the provided path,
// and used to set the current environment.
// It is assumed that all parameters are of type "SecureString".
//
// Deprecated: use SetEnvVarsUsingSSM
func Parse(path string) error {
	return SetEnvVarsUsingSSM(path)
}

// SetEnvVarsUsingSSM sets environment variables using AWS SSM.
//
// All parameters are fetched from SSM using the provided path,
// and used to set the current environment.
// It is assumed that all parameters are of type "SecureString".
func SetEnvVarsUsingSSM(path string) error {

	if path == "" {
		return nil
	}

	if !strings.HasSuffix(path, "/") {
		path += "/"
	}

	awsConfig, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return err
	}

	// ssmClient is a new SSM session for interacting with SSM.
	var ssmClient = ssm.NewFromConfig(awsConfig)

	input := ssm.GetParametersByPathInput{
		Path:           aws.String(path),
		WithDecryption: aws.Bool(true),
	}

	// Paginators are how pagination is now done in github.com/aws/aws-sdk-go-v2 style
	// See: https://docs.aws.amazon.com/sdk-for-go/v2/developer-guide/migrate-gosdk.html
	paginator := ssm.NewGetParametersByPathPaginator(ssmClient, &input)

	for paginator.HasMorePages() {

		paginatorOutput, err := paginator.NextPage(context.Background())
		if err != nil {
			return err
		}

		for _, param := range paginatorOutput.Parameters {

			// Skip empty secrets
			if param.Name == nil || param.Value == nil {
				continue
			}

			envVarName := strings.TrimPrefix(*param.Name, path)

			// Skip obviously invalid environment variables
			if envVarName == "" || *param.Value == "" {
				continue
			}

			err = os.Setenv(envVarName, *param.Value)
			if err != nil {
				return err
			}

		}

	}

	return nil
}
