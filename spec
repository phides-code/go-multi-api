let's make a plan to set up a Go aws lambda that will act as an API doing CRUD on a dynamodb table. 

the table name is "bananas"

each item is a "banana" of this shape:

{
    id: string;
    content: string;
}

the lambda will handle incoming http requests and accept GET, PUT, POST, and DELETE methods.

the url param will be /bananas/ (for now, other paths to be added in the future). invalid url params will be blocked by the api gateway before reaching the lambda. 

GET, PUT and DELETE accept /bananas/{id} as a url param.

GET with no id param means get all.

POST generates a uuid as the id. 

it should handle client errors (invalid id) and internal server errors. 

it should be done in uncle Bob style clean code, DRY, SOLID, and easy to extend for future tables or changes to the entity types.  

don't worry about setting up lambda template or aws infrastructure for now. 

what other info do i need to furnish?
==

API Gateway integration shape: let's use events.APIGatewayProxyRequest / APIGatewayProxyResponse.

DynamoDB key design: PK = id

PUT semantics: Update only. 404 if missing. id is found only in the path. 

POST request body: "content" only. validation: "content" required, must be a string between 1 and 1000 chars. 

GET all behavior: Paginated (limit + cursor/LastEvaluatedKey) but no query params for now. 

Response contract: always respond with this shape: 

{
    data: Banana or Banana[] or nil
    error: string or nil 
}

always Content-Type: application/json

ID format and validation: POST generates UUID v4, GET/PUT/DELETE {id} should validate id as a uuid. 

DELETE behavior: Hard delete from DynamoDB, respond with 200 and the full object that was just deleted.

Auth: None for now.

CORS: will be handled by api gateway. 

Logging: Structured JSON logs.

Table name: Hardcode "bananas".

AWS region: i don't think is needed yet? it will be known by the build script. 

Go version: the latest stable for lambda. 

Dependencies: AWS SDK v2 (github.com/aws/aws-sdk-go-v2) yes 

Extensibility preferences: One Lambda, router dispatches /bananas, /apples, … (monolith Lambda)
shared middleware (logging, error handling) should live in a small internal/platform package.

testing: Unit tests with mocked repository.

