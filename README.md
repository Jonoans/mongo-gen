# mongo-gen

mongo-gen is a Go library for generating Mongo models for Go programs.

In your project, provide a `orm.yml` file with the following information:
- The input package's name and path containing your input structs.
- The output package name and path to output the structs and its relevant methods.
- Run `go run github.com/jonoans/mongo-gen generate`

mongo-gen attempts to generate working (hopefully) methods to resolve references to other collections.

Check out `orm.yml` for an example configuration and the `examples` directory for the input models and generated code.

## Input Models

Provide your input models as structs in the package indicated in the `orm.yml` file.
- Only structs which have the [`codegen.BaseModel`](https://github.com/Jonoans/mongo-gen/blob/main/codegen/base_model.go) field embedded are recognised as collection documents.

## Output Models

The output models will contain additional methods to hopefully make life easier.
- `GetResolved_[FIELD NAME]` method for automatically resolving references.
- `Queried`, `Creating`, `Created`, `Saving`, `Saved`, `Updating`, `Updated`, `Deleting`, `Deleted` hook methods.

## codegen_.go

Included in the generated files, contains functions for using models.
- API is similar to [https://github.com/Kamva/mgm](https://github.com/Kamva/mgm)