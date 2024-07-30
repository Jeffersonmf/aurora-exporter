# go-aurora-exporter

#### This project consists of creating, managing and importing data currently in RedShift to the new hot bases in RDS Aurora.


![](./golang.png)

#### The entire core was built in the Go language, aiming to be fast, resilient, and easy to maintain

![](./migrate1.png)

## Getting Started #
In this tutorial, you will learn how to use the migrate tool (written in Go and quite popular in Golang community) to execute database migrations. As a second part, you will write some Go code to read the data from the database.

You will use PostgreSQL as the database of choice in this article, but migrate is compatible with many more databases check out the list here.

![](./migrate3.png)

## Getting started

To make it easy for you to get started with GitLab, here's a list of recommended next steps.

Already a pro? Just edit this README.md and make it your own. Want to make it easy? [Use the template at the bottom](#editing-this-readme)!

### Setup 

#### Installation 

migrate is a cli tool that is used to run migrations, it can also be used programmatically, but in this tutorial we are going to use it via cli. There are multiple ways to install the migrate cli, such as brew, scoop, etc. Take a look at the installation document to see all the available options.

We are going to install it using go install.

```
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

#### Database 

To run the sample Postgres database, I have provided you with a docker-compose.yaml file. Clone the github project and run docker compose up. This will run a Postgres database on port **5454**.

You are free to setup a Postgres DB as you like. There is nothing special in the above setup.

### Creating Golang Migrations 

You will create a posts table which will have to 2 columns, title and body.

To create a new migration run the migrate create command with appropriate options.

```
migrate create -ext sql -dir db/migrations create_posts_table
```

**ext** specifies the file extension to use when creating migrations file.

**dir** specifies which directory to create the migrations in.
This will create two migrations in db/migrations folder matching the pattern: <timestamp>_create_posts_table.down.sql and <timestamp>_create_posts_table.up.sql.

* UP migration will contain the sql to create the post table.
* DOWN migration will contain the sql to revert what has been done in the up migration.

### Writing SQL 

In the <timestamp>_create_posts_table.up.sql migration file write the sql to create posts table.

```
CREATE TABLE IF NOT EXISTS posts (title varchar, body varchar);
```

In the <timestamp>_create_posts_table.down.sql migration file write the sql to drop posts table.
```
DROP TABLE IF EXISTS posts;
```

### Running migrations 

**migrate** needs a way to connect the database to execute the sql statements. For this you will need a valid postgres connection string following the format:

```
postgres://<username>:<password>@localhost:<port>/<db_name>?sslmode=disable
```

If you are using the database using the docker-compose.yaml method specified above, the connection string will look like this:

```
postgres://postgres:postgres@localhost:5454/postgres?sslmode=disable
```

To run migrations use migrate up command with the appropriate options.

```
export DB_URL='postgres://postgres:postgres@localhost:5454/postgres?sslmode=disable'
```

# Run migrations
migrate -database ${DB_URL} -path db/migrations up

![](./migrate2.png)

## Add your files

- [ ] [Create](https://docs.gitlab.com/ee/user/project/repository/web_editor.html#create-a-file) or [upload](https://docs.gitlab.com/ee/user/project/repository/web_editor.html#upload-a-file) files
- [ ] [Add files using the command line](https://docs.gitlab.com/ee/gitlab-basics/add-file.html#add-a-file-using-the-command-line) or push an existing Git repository with the following command:

```
cd existing_repo
git remote add origin https://new.gitlab.com/data/etl/go-aurora-exporter.git
git branch -M main
git push -uf origin main
```

## Integrate with your tools

- [ ] [Set up project integrations](https://new.gitlab.com/data/etl/go-aurora-exporter/-/settings/integrations)

## Collaborate with your team

- [ ] [Invite team members and collaborators](https://docs.gitlab.com/ee/user/project/members/)
- [ ] [Create a new merge request](https://docs.gitlab.com/ee/user/project/merge_requests/creating_merge_requests.html)
- [ ] [Automatically close issues from merge requests](https://docs.gitlab.com/ee/user/project/issues/managing_issues.html#closing-issues-automatically)
- [ ] [Enable merge request approvals](https://docs.gitlab.com/ee/user/project/merge_requests/approvals/)
- [ ] [Set auto-merge](https://docs.gitlab.com/ee/user/project/merge_requests/merge_when_pipeline_succeeds.html)

## Test and Deploy

Use the built-in continuous integration in GitLab.

- [ ] [Get started with GitLab CI/CD](https://docs.gitlab.com/ee/ci/quick_start/index.html)
- [ ] [Analyze your code for known vulnerabilities with Static Application Security Testing (SAST)](https://docs.gitlab.com/ee/user/application_security/sast/)
- [ ] [Deploy to Kubernetes, Amazon EC2, or Amazon ECS using Auto Deploy](https://docs.gitlab.com/ee/topics/autodevops/requirements.html)
- [ ] [Use pull-based deployments for improved Kubernetes management](https://docs.gitlab.com/ee/user/clusters/agent/)
- [ ] [Set up protected environments](https://docs.gitlab.com/ee/ci/environments/protected_environments.html)

***

# Editing this README

When you're ready to make this README your own, just edit this file and use the handy template below (or feel free to structure it however you want - this is just a starting point!). Thanks to [makeareadme.com](https://www.makeareadme.com/) for this template.

## Suggestions for a good README

Every project is different, so consider which of these sections apply to yours. The sections used in the template are suggestions for most open source projects. Also keep in mind that while a README can be too long and detailed, too long is better than too short. If you think your README is too long, consider utilizing another form of documentation rather than cutting out information.

## Name
Choose a self-explaining name for your project.

## Description
Let people know what your project can do specifically. Provide context and add a link to any reference visitors might be unfamiliar with. A list of Features or a Background subsection can also be added here. If there are alternatives to your project, this is a good place to list differentiating factors.

## Badges
On some READMEs, you may see small images that convey metadata, such as whether or not all the tests are passing for the project. You can use Shields to add some to your README. Many services also have instructions for adding a badge.

## Visuals
Depending on what you are making, it can be a good idea to include screenshots or even a video (you'll frequently see GIFs rather than actual videos). Tools like ttygif can help, but check out Asciinema for a more sophisticated method.

## Installation
Within a particular ecosystem, there may be a common way of installing things, such as using Yarn, NuGet, or Homebrew. However, consider the possibility that whoever is reading your README is a novice and would like more guidance. Listing specific steps helps remove ambiguity and gets people to using your project as quickly as possible. If it only runs in a specific context like a particular programming language version or operating system or has dependencies that have to be installed manually, also add a Requirements subsection.

## Usage
Use examples liberally, and show the expected output if you can. It's helpful to have inline the smallest example of usage that you can demonstrate, while providing links to more sophisticated examples if they are too long to reasonably include in the README.

## Support
Tell people where they can go to for help. It can be any combination of an issue tracker, a chat room, an email address, etc.

## Roadmap
If you have ideas for releases in the future, it is a good idea to list them in the README.

## Contributing
State if you are open to contributions and what your requirements are for accepting them.

For people who want to make changes to your project, it's helpful to have some documentation on how to get started. Perhaps there is a script that they should run or some environment variables that they need to set. Make these steps explicit. These instructions could also be useful to your future self.

You can also document commands to lint the code or run tests. These steps help to ensure high code quality and reduce the likelihood that the changes inadvertently break something. Having instructions for running tests is especially helpful if it requires external setup, such as starting a Selenium server for testing in a browser.

## Authors and acknowledgment
Show your appreciation to those who have contributed to the project.

## License
For open source projects, say how it is licensed.

## Project status
If you have run out of energy or time for your project, put a note at the top of the README saying that development has slowed down or stopped completely. Someone may choose to fork your project or volunteer to step in as a maintainer or owner, allowing your project to keep going. You can also make an explicit request for maintainers.
