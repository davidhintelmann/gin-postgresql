# gin-postgresql

Using the [gin webframework](https://gin-gonic.com/) for [go lang](https://go.dev) and [PostgreSQL](https://www.postgresql.org/) database for creating a simple RESTful API

Please note in `connect/password_edit.go` there is an `ImportPassword_()` function which needs to edited. The underscore needs to be removed at the end of the function's name and
the filename needs to be renamed to `password.go`. Then enter your PostgreSQL password into the function's return statement.

For starters I recommend following [this tutorial](https://go.dev/doc/tutorial/web-service-gin) from go lang for using gin web framework for creating a simple RESTful API.
