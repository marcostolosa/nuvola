package connector

import (
	"encoding/json"
	"fmt"
	"nuvola/connector/services/aws/database"
	"nuvola/connector/services/aws/ec2"
	"nuvola/connector/services/aws/iam"
	"nuvola/connector/services/aws/lambda"
	"nuvola/connector/services/aws/s3"
	neo4jconnector "nuvola/connector/services/neo4j"
	cli_output "nuvola/tools/cli/output"
	nuvolaerror "nuvola/tools/error"
	"os"
	"regexp"

	"github.com/joho/godotenv"
)

func NewStorageConnector() *StorageConnector {
	// Load .env
	if err := godotenv.Load(); err != nil {
		nuvolaerror.HandleError(err, "NewStorageConnector", "Error loading .env file")
	}
	neo4jURL := os.Getenv("NEO4J_URL")
	neo4jUsername := "neo4j"
	neo4jPassword := os.Getenv("PASSWORD")
	client, err := neo4jconnector.Connect(neo4jURL, neo4jUsername, neo4jPassword)
	if err != nil {
		nuvolaerror.HandleError(err, "NewStorageConnector", "Error connecting to database")
	}
	connector := &StorageConnector{
		Client: *client,
	}
	return connector
}

func (sc *StorageConnector) FlushAll() *StorageConnector {
	sc.Client.DeleteAll()
	return sc
}

func (sc *StorageConnector) ImportResults(what string, content []byte) {
	var whoami = regexp.MustCompile(`^Whoami`)
	var credentialReport = regexp.MustCompile(`^CredentialReport`)
	var users = regexp.MustCompile(`^Users`)
	var groups = regexp.MustCompile(`^Groups`)
	var roles = regexp.MustCompile(`^Roles`)
	var buckets = regexp.MustCompile(`^Buckets`)
	var ec2s = regexp.MustCompile(`^EC2s`)
	var vpcs = regexp.MustCompile(`^VPCs`)
	var lambdas = regexp.MustCompile(`^Lambdas`)
	var rds = regexp.MustCompile(`^RDS`)
	var dynamodbs = regexp.MustCompile(`^DynamoDBs`)
	var redshiftdbs = regexp.MustCompile(`^RedshiftDBs`)

	cli_output.PrintDarkGreen(fmt.Sprintf("Importing: %s", what))
	switch {
	case whoami.MatchString(what):
	case credentialReport.MatchString(what):
	case users.MatchString(what):
		contentStruct := []iam.User{}
		_ = json.Unmarshal(content, &contentStruct)
		sc.Client.AddUsers(&contentStruct)
	case groups.MatchString(what):
		contentStruct := []iam.Group{}
		_ = json.Unmarshal(content, &contentStruct)
		sc.Client.AddGroups(&contentStruct)
	case roles.MatchString(what):
		contentStruct := []iam.Role{}
		_ = json.Unmarshal(content, &contentStruct)
		sc.Client.AddRoles(&contentStruct)
		sc.Client.AddLinksToResourcesIAM()
	case buckets.MatchString(what):
		contentStruct := []s3.Bucket{}
		_ = json.Unmarshal(content, &contentStruct)
		sc.Client.AddBuckets(&contentStruct)
	case ec2s.MatchString(what):
		contentStruct := []ec2.Instance{}
		_ = json.Unmarshal(content, &contentStruct)
		sc.Client.AddEC2(&contentStruct)
	case vpcs.MatchString(what):
		contentStruct := ec2.VPC{}
		_ = json.Unmarshal(content, &contentStruct)
		sc.Client.AddVPC(&contentStruct)
	case lambdas.MatchString(what):
		contentStruct := []lambda.Lambda{}
		_ = json.Unmarshal(content, &contentStruct)
		sc.Client.AddLambda(&contentStruct)
	case rds.MatchString(what):
		contentStruct := database.RDS{}
		_ = json.Unmarshal(content, &contentStruct)
		sc.Client.AddRDS(&contentStruct)
	case dynamodbs.MatchString(what):
		contentStruct := []database.DynamoDB{}
		_ = json.Unmarshal(content, &contentStruct)
		sc.Client.AddDynamoDB(&contentStruct)
	case redshiftdbs.MatchString(what):
		contentStruct := []database.RedshiftDB{}
		_ = json.Unmarshal(content, &contentStruct)
		sc.Client.AddRedshift(&contentStruct)
	default:
		nuvolaerror.HandleError(nil, "ImportResults", "Error importing data")
	}
}

func (sc *StorageConnector) ImportBulkResults(content map[string]interface{}) {
	for k, v := range content {
		value, err := json.Marshal(v)
		if err != nil {
			nuvolaerror.HandleError(err, "ImportBulkResults", "Error on marshalling data")
		}
		sc.ImportResults(k, value)
	}
}

func (sc *StorageConnector) Query(query string, arguments map[string]interface{}) []map[string]interface{} {
	return sc.Client.Query(query, arguments)
}
