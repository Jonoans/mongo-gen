package input

import "github.com/jonoans/mongo-gen/codegen"

type SubModel struct {
}

type AnotherModel struct {
	codegen.BaseModel
	Sub SubModel `bson:"sub"`
}

type Model struct {
	codegen.BaseModel
	Sub                   SubModel
	Reference             AnotherModel
	ReferencePtr          *AnotherModel
	ReferenceSlice        []AnotherModel
	ReferenceSliceInSlice [][]*AnotherModel
	ReferenceMap          map[string]AnotherModel
	ReferenceMapPtr       map[string]*AnotherModel
	ReferencePtrSlice     *[]AnotherModel
	ReferencePtrMap       *map[string]AnotherModel
}
