package user

import (
	"encoding/json"
	"errors"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/zzyong99/Go-Serverless/pkg/validators"
)

var (
	ErrorFailedToUnmarshalRecord = "failed to unmarshal record"
	ErrorFailedToFetchRecord     = "failed to fetch record"
	ErrorInvalidUserDara         = "invalid user data"
	ErrorInvalidEmail            = ""
	ErrorCouldNotMarshalItem     = ""
	ErrorCouldNotDeleteItem      = ""
	ErrorCouldNotDynamoPutItem   = ""
	ErrorUserAlreadyExist        = ""
	ErrorUSerDoesNotExist        = ""
)

type User struct {
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

func FetchUser(email, tableName string, dynaClient dynamodbiface.DynamoDBAPI) (*User, error) {
	input := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"email": {
				S: aws.String(email),
			},
		},
		TableName: aws.String(tableName),
	}

	result, err := dynaClient.GetItem(input)
	if err != nil {
		return nil, errors.New(ErrorFailedToFetchRecord)
	}

	item := new(User)
	err = dynamodbattribute.UnmarshalMap(result.Item, item)
	if err != nil {
		return nil, errors.New(ErrorCouldNotMarshalItem)
	}
	return item, nil
}

func FetchUsers(tableName string, dynaClient dynamodbiface.DynamoDBAPI) (*[]User, error) {
	input := &dynamodb.ScanInput{
		TableName: aws.String(tableName),
	}

	result, err := dynaClient.Scan(input)
	if err != nil {
		return nil, errors.New(ErrorFailedToFetchRecord)
	}
	item := new([]User)
	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, item)
	if err != nil {
		return nil, errors.New(ErrorCouldNotMarshalItem)
	}
	return item, nil

}

func CreateUser(req events.APIGatewayProxyRequest, tableName string, dynaClient dynamodbiface.DynamoDBAPI) (*User, error) {
	var user User

	if err := json.Unmarshal([]byte(req.Body), &user); err != nil { //
		return nil, errors.New(ErrorInvalidUserDara)
	}
	if !validators.IsEmailValid(user.Email) {
		return nil, errors.New(ErrorInvalidEmail)
	}

	currentUser, _ := FetchUser(user.Email, tableName, dynaClient) //
	if currentUser != nil && len(currentUser.Email) != 0 {
		return nil, errors.New(ErrorUserAlreadyExist)
	}

	av, err := dynamodbattribute.MarshalMap(user) //
	if err != nil {
		return nil, errors.New(ErrorCouldNotMarshalItem)
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(tableName),
	}

	_, err = dynaClient.PutItem(input)
	if err != nil {
		return nil, errors.New(ErrorCouldNotDynamoPutItem)
	}
	return &user, nil
}

func UpdateUser(req events.APIGatewayProxyRequest, tableName string, dynaClient dynamodbiface.DynamoDBAPI) (*User, error) {
	var user User
	if err := json.Unmarshal([]byte(req.Body), &user); err != nil {
		return nil, errors.New(ErrorInvalidEmail)
	}

	currentUser, _ := FetchUser(user.Email, tableName, dynaClient)
	if currentUser != nil && len(currentUser.Email) == 0 {
		return nil, errors.New(ErrorUSerDoesNotExist)
	}

	av, err := dynamodbattribute.MarshalMap(user)
	if err != nil {
		return nil, errors.New(ErrorCouldNotMarshalItem)
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(tableName),
	}

	_, err = dynaClient.PutItem(input)
	if err != nil {
		return nil, errors.New(ErrorCouldNotDynamoPutItem)
	}

	return &user, nil
}

func DeleteUser(req events.APIGatewayProxyRequest, tableName string, dynaClient dynamodbiface.DynamoDBAPI) error {
	email := req.QueryStringParameters["email"]
	input := &dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"email": {
				S: aws.String(email),
			},
		},
		TableName: aws.String(tableName),
	}

	_, err := dynaClient.DeleteItem(input)
	if err != nil {
		return errors.New(ErrorCouldNotDeleteItem)
	}
	return nil
}
