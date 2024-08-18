# forum

### last update 1.4 (18.08.2024)
###### see changes in Changelog.txt

User can register, login, comment, create topics and like posts. Users cant remove posts or topics they created, but its possible to edit posts until a certain time.
All data is stored in the database (sqlite3)
Forum has 4 main categories with different topics. Main categories cant be edited by users.

#### Dockerfile included.
1) ##### Create an image 
> docker build -t YOUR-IMAGE-NAME .
2) ##### Create and run the container
> docker run --name=YOUR-CONTAINER-NAME -p YOUR-PREFERABLE-PORT:8080 YOUR-IMAGE-NAME
3) ##### Run server using port YOUR-PREFERABLE-PORT

##### To stop Docker container: 
> docker stop YOUR-CONTAINER-NAME
##### To start again Docker container: 
> docker start YOUR-CONTAINER-NAME
##### To delete Docker container forever: 
> docker rm YOUR-CONTAINER-NAME
##### To delete a Docker image forever: 
> docker rmi YOUR-IMAGE-NAME

Planning that work on the project will continue
The project will be updated and all changes will be written here or in a text file attached with project files.