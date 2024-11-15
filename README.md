# forum
#### Go Web App
### Current version: 2.1 (14.11.2024)
#### see changes in Changelog.txt

<p>Users can sign up, login, comment, create topics, like posts, edit posts they sent, and modify their password on My Profile page.<br>
All the data is stored in the database (SQLite3)<br>
Forum has main categories with different topics. Main categories cant be edited by users.<br>
Forum has groups for users. Standard group - users.<br>
Admins can delete/edit threads and posts, edit passwords, promote and demote users</p>

##### To run app server: 
> go run .
##### or
> go run server.go

#### Dockerfile is included.
1) ##### Create an image
> docker build -t YOUR-IMAGE-NAME .
2) ##### Create and run the container
> docker run --name=YOUR-CONTAINER-NAME -p YOUR-PREFERABLE-PORT:8080 YOUR-IMAGE-NAME
3) ##### Run server using port [YOUR-PREFERABLE-PORT]

##### To stop Docker container: 
> docker stop YOUR-CONTAINER-NAME
##### To start again Docker container: 
> docker start YOUR-CONTAINER-NAME
##### To delete Docker container forever: 
> docker rm YOUR-CONTAINER-NAME
##### To delete a Docker image forever: 
> docker rmi YOUR-IMAGE-NAME

<p>I plan to continue working on the project.<br>
The project will be updated, and all changes will be documented either here or in a text file included with the project files.</p>