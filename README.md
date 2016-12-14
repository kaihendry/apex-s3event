Derived from <https://github.com/apex/apex/tree/master/_examples/go>

To run the example first setup your [AWS
Credentials](http://apex.run/#aws-credentials), and ensure "role" in
./project.json is set to your Lambda function ARN.

or `rm project.json && apex init`, but be sure to set the S3URI environment and
add S3 write access to your newly created role.

Deploy the functions:

```
$ apex deploy
```

Try it out:

```
$ apex invoke S3event < event.json
```
