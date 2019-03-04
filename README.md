# Overview

The purpose of this [POC](https://en.wikipedia.org/wiki/Proof_of_concept) is to showcase using [Amazon DocumentDB (with MongoDB compatibility)](https://aws.amazon.com/documentdb/) with [AWS Fargate](https://aws.amazon.com/fargate/) to perform batch compute operations on data persisted in the database.

**Note:** You must have the new [ARN format and resource ID format](https://aws.amazon.com/ecs/faqs/#Transition_to_new_ARN_and_ID_format) enabled before launching the [AWS CloudFormation](https://aws.amazon.com/cloudformation/) template in order to support tags ([blog post](https://aws.amazon.com/blogs/compute/migrating-your-amazon-ecs-deployment-to-the-new-arn-and-resource-id-format-2/)).

To test the job manager locally (this repo), set the following environment variables (must match your deployed cluster):
```
export DOCUMENT_DB_ENDPOINT=localhost
export DOCUMENT_DB_PORT=27017
export DOCUMENT_DB_USER=test
export DOCUMENT_DB_PASSWORD=test
export DOCUMENT_DB_PEM=/FULL_PATH/fargate-documentdb-compute-poc/local.pem
export LOCAL=true
```

To test locally with MongoDB 3.6.9, you can use the following commands:

Run the database:
```
mongod --sslMode requireSSL --sslPEMKeyFile /FULL_PATH/fargate-documentdb-compute-poc/local.pem
```

Access the database locally via the shell:

```
mongo --host localhost --port 27017 --ssl --sslPEMKeyFile=/FULL_PATH/fargate-documentdb-compute-poc/local.pem --sslAllowInvalidCertificates --sslAllowInvalidHostnames
```

