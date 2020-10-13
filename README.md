## swaggerdoc

Creates [Swagger](https://swagger.io) definition files in both json and yaml
as well as a set of standalone documentation files.

The documentation is extracted by parsing your Go source using
[Swag](https://github.com/swaggo/swag). Please refer to the documentation
there for how to add the proper directives to your source code.

Note that this project uses an embedded copy of the swag project, stripped
down to the parts needed by this project. This was done to get around an
issue with swag indirectly depending upon a non-existent package due to an
upstream tag being removed. At some point in the future after this has been
resolved, I may revert this change back.
