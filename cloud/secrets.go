package cloud

import (
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	log "github.com/golang/glog"
)

type AppSecrets struct {
	DatabasePassword string
}

var Secrets *AppSecrets

func init() {

	Secrets = &AppSecrets{}

	// Get from a local env var or pull from secrets manager
	if len(os.Getenv("DOCUMENT_DB_LOCAL")) > 0 {
		Secrets.DatabasePassword = os.Getenv("DOCUMENT_DB_PASSWORD")
		return
	}

	if databasePassword, err := loadSecret(os.Getenv("DOCUMENT_DB_PASSWORD_SECRET_NAME")); err == nil {
		Secrets.DatabasePassword = databasePassword
	} else {
		log.Errorf("Cannot load secret: %s - problem: %v", os.Getenv("DOCUMENT_DB_PASSWORD_SECRET_NAME"), err)
	}
}

func loadSecret(secretName string) (string, error) {
	svc := secretsmanager.New(session.New())
	if result, err := svc.GetSecretValue(&secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(secretName),
		VersionStage: aws.String("AWSCURRENT"),
	}); err != nil {
		return "", err
	} else {
		return *result.SecretString, nil
	}
}
