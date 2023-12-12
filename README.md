## Instructions

* first you need to create a policy using aws IAM service.

* Then create a .env file and add all variables mentioned in the .env.example file

* Note: before automate this, you should only have less than 5 versions in your aws policy versions. other than that it will throw an error from aws side.

* you can add a new bucket name that is not existing inside your S3 bucket or existing bucket name. Both are acceptable.

* After configured all,  run this command to download go modules..
    ```bash
        go mod tidy
    ```

* Then run the application using below command.
    ```bash
        go run main.go
    ```

* for upload an object (POST request)
    ```bash
        curl --location 'http://localhost:3000/aws_upload' \
        --form 'file=@"/images/test_image.png"'
    ```

* To get presigned url (POST request)
    ```bash
        curl --location --request POST 'http://localhost:3000/aws_presigned_url'
    ```

* To delete existing object (DELETE request)
    ```bash
        curl --location --request DELETE 'http://localhost:3000/aws_delete_object'
    ```