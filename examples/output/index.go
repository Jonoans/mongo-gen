package output

import (
	"context"

	"github.com/jonoans/mongo-gen/codegen"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ATestStruct struct {
}

type AnotherModel struct {
	codegen.BaseModel
	Sub SubModel `bson:"sub"`
}

type Model struct {
	codegen.BaseModel
	Sub                   SubModel
	Reference             primitive.ObjectID
	ReferencePtr          *primitive.ObjectID
	ReferenceSlice        []primitive.ObjectID
	ReferenceSliceInSlice [][]*primitive.ObjectID
	ReferenceMap          map[string]primitive.ObjectID
	ReferenceMapPtr       map[string]*primitive.ObjectID
	ReferencePtrSlice     *[]primitive.ObjectID
	ReferencePtrMap       *map[string]primitive.ObjectID

	errReference                  error
	initReference                 bool
	resolvedReference             AnotherModel
	errReferencePtr               error
	initReferencePtr              bool
	resolvedReferencePtr          *AnotherModel
	errReferenceSlice             error
	initReferenceSlice            bool
	resolvedReferenceSlice        []AnotherModel
	errReferenceSliceInSlice      error
	initReferenceSliceInSlice     bool
	resolvedReferenceSliceInSlice [][]*AnotherModel
	errReferenceMap               error
	initReferenceMap              bool
	resolvedReferenceMap          map[string]AnotherModel
	errReferenceMapPtr            error
	initReferenceMapPtr           bool
	resolvedReferenceMapPtr       map[string]*AnotherModel
	errReferencePtrSlice          error
	initReferencePtrSlice         bool
	resolvedReferencePtrSlice     *[]AnotherModel
	errReferencePtrMap            error
	initReferencePtrMap           bool
	resolvedReferencePtrMap       *map[string]AnotherModel
}

type SubModel struct {
}

func (*AnotherModel) CollectionName() string {
	return "anotherModel"
}

func (*Model) CollectionName() string {
	return "model"
}

func (m *AnotherModel) Queried() error {
	return nil
}

func (m *AnotherModel) Creating() error {
	return nil
}

func (m *AnotherModel) Created() error {
	return nil
}

func (m *AnotherModel) Saving() error {
	return nil
}

func (m *AnotherModel) Saved() error {
	return nil
}

func (m *AnotherModel) Updating() error {
	return nil
}

func (m *AnotherModel) Updated() error {
	return nil
}

func (m *AnotherModel) Deleting() error {
	return nil
}

func (m *AnotherModel) Deleted() error {
	return nil
}

func (m *Model) Queried() error {
	return nil
}

func (m *Model) Creating() error {
	return nil
}

func (m *Model) Created() error {
	return nil
}

func (m *Model) Saving() error {
	return nil
}

func (m *Model) Saved() error {
	return nil
}

func (m *Model) Updating() error {
	return nil
}

func (m *Model) Updated() error {
	return nil
}

func (m *Model) Deleting() error {
	return nil
}

func (m *Model) Deleted() error {
	return nil
}

func (m *Model) GetResolved_Reference() (AnotherModel, error) {
	if m.initReference {
		return m.resolvedReference, m.errReference
	}
	m.errReference = FindByObjectID(&m.resolvedReference, m.Reference)
	m.initReference = true
	return m.resolvedReference, m.errReference
}

func (m *Model) GetResolved_ReferencePtr() (*AnotherModel, error) {
	if m.initReferencePtr {
		return m.resolvedReferencePtr, m.errReferencePtr
	}
	if m.ReferencePtr == nil {
		m.initReferencePtr = true
		return m.resolvedReferencePtr, m.errReferencePtr
	}
	m.resolvedReferencePtr = new(AnotherModel)
	m.errReferencePtr = FindByObjectID(m.resolvedReferencePtr, m.ReferencePtr)
	m.initReferencePtr = true
	return m.resolvedReferencePtr, m.errReferencePtr
}

func (m *Model) GetResolved_ReferenceSlice() ([]AnotherModel, error) {
	if m.initReferenceSlice {
		return m.resolvedReferenceSlice, m.errReferenceSlice
	}
	if m.ReferenceSlice == nil {
		m.initReferenceSlice = true
		return m.resolvedReferenceSlice, m.errReferenceSlice
	}
	m.resolvedReferenceSlice = make([]AnotherModel, 0)
	m.errReferenceSlice = FindByObjectIDs(&m.resolvedReferenceSlice, m.ReferenceSlice)
	m.initReferenceSlice = true
	return m.resolvedReferenceSlice, m.errReferenceSlice
}

func (m *Model) GetResolved_ReferenceSliceInSlice() ([][]*AnotherModel, error) {
	if m.initReferenceSliceInSlice {
		return m.resolvedReferenceSliceInSlice, m.errReferenceSliceInSlice
	}
	if m.ReferenceSliceInSlice == nil {
		m.initReferenceSliceInSlice = true
		return m.resolvedReferenceSliceInSlice, m.errReferenceSliceInSlice
	}
	m.resolvedReferenceSliceInSlice = make([][]*AnotherModel, len(m.ReferenceSliceInSlice))
	for ka, va := range m.ReferenceSliceInSlice {
		if va == nil {
			continue
		}
		m.resolvedReferenceSliceInSlice[ka] = make([]*AnotherModel, len(va))
		for kb, vb := range va {
			if vb == nil {
				continue
			}
			m.resolvedReferenceSliceInSlice[ka][kb] = new(AnotherModel)
			m.errReferenceSliceInSlice = FindByObjectID(m.resolvedReferenceSliceInSlice[ka][kb], vb)
			if m.errReferenceSliceInSlice != nil {
				m.initReferenceSliceInSlice = true
				return m.resolvedReferenceSliceInSlice, m.errReferenceSliceInSlice
			}
		}
		if m.errReferenceSliceInSlice != nil {
			m.initReferenceSliceInSlice = true
			return m.resolvedReferenceSliceInSlice, m.errReferenceSliceInSlice
		}
	}
	m.initReferenceSliceInSlice = true
	return m.resolvedReferenceSliceInSlice, m.errReferenceSliceInSlice
}

func (m *Model) GetResolved_ReferenceMap() (map[string]AnotherModel, error) {
	if m.initReferenceMap {
		return m.resolvedReferenceMap, m.errReferenceMap
	}
	if m.ReferenceMap == nil {
		m.initReferenceMap = true
		return m.resolvedReferenceMap, m.errReferenceMap
	}
	for ka, va := range m.ReferenceMap {
		bAssign := AnotherModel{}
		m.errReferenceMap = FindByObjectID(&bAssign, va)
		m.resolvedReferenceMap[ka] = bAssign
		if m.errReferenceMap != nil {
			m.initReferenceMap = true
			return m.resolvedReferenceMap, m.errReferenceMap
		}
	}
	m.initReferenceMap = true
	return m.resolvedReferenceMap, m.errReferenceMap
}

func (m *Model) GetResolved_ReferenceMapPtr() (map[string]*AnotherModel, error) {
	if m.initReferenceMapPtr {
		return m.resolvedReferenceMapPtr, m.errReferenceMapPtr
	}
	if m.ReferenceMapPtr == nil {
		m.initReferenceMapPtr = true
		return m.resolvedReferenceMapPtr, m.errReferenceMapPtr
	}
	for ka, va := range m.ReferenceMapPtr {
		if va == nil {
			m.resolvedReferenceMapPtr[ka] = nil
			continue
		}
		m.resolvedReferenceMapPtr[ka] = new(AnotherModel)
		bAssign := new(AnotherModel)
		m.errReferenceMapPtr = FindByObjectID(bAssign, va)
		m.resolvedReferenceMapPtr[ka] = bAssign
		if m.errReferenceMapPtr != nil {
			m.initReferenceMapPtr = true
			return m.resolvedReferenceMapPtr, m.errReferenceMapPtr
		}
	}
	m.initReferenceMapPtr = true
	return m.resolvedReferenceMapPtr, m.errReferenceMapPtr
}

func (m *Model) GetResolved_ReferencePtrSlice() (*[]AnotherModel, error) {
	if m.initReferencePtrSlice {
		return m.resolvedReferencePtrSlice, m.errReferencePtrSlice
	}
	if m.ReferencePtrSlice == nil {
		m.initReferencePtrSlice = true
		return m.resolvedReferencePtrSlice, m.errReferencePtrSlice
	}
	m.resolvedReferencePtrSlice = new([]AnotherModel)
	aAssign := *m.resolvedReferencePtrSlice
	aID := *m.ReferencePtrSlice
	if aID == nil {
		m.initReferencePtrSlice = true
		return m.resolvedReferencePtrSlice, m.errReferencePtrSlice
	}
	m.errReferencePtrSlice = FindByObjectIDs(&aAssign, aID)
	m.initReferencePtrSlice = true
	return m.resolvedReferencePtrSlice, m.errReferencePtrSlice
}

func (m *Model) GetResolved_ReferencePtrMap() (*map[string]AnotherModel, error) {
	if m.initReferencePtrMap {
		return m.resolvedReferencePtrMap, m.errReferencePtrMap
	}
	if m.ReferencePtrMap == nil {
		m.initReferencePtrMap = true
		return m.resolvedReferencePtrMap, m.errReferencePtrMap
	}
	m.resolvedReferencePtrMap = new(map[string]AnotherModel)
	aAssign := *m.resolvedReferencePtrMap
	aID := *m.ReferencePtrMap
	if aID == nil {
		m.initReferencePtrMap = true
		return m.resolvedReferencePtrMap, m.errReferencePtrMap
	}
	for kb, vb := range aID {
		cAssign := AnotherModel{}
		m.errReferencePtrMap = FindByObjectID(&cAssign, vb)
		aAssign[kb] = cAssign
		if m.errReferencePtrMap != nil {
			m.initReferencePtrMap = true
			return m.resolvedReferencePtrMap, m.errReferencePtrMap
		}
	}
	m.initReferencePtrMap = true
	return m.resolvedReferencePtrMap, m.errReferencePtrMap
}

func (m *AnotherModel) AggregateFirst(pipeline interface{}, opts ...*options.AggregateOptions) (bool, error) {
	return AggregateFirst(m, pipeline, opts...)
}

func (m *AnotherModel) Find(query interface{}, opts ...*options.FindOneOptions) error {
	return FindOne(m, query, opts...)
}

func (m *AnotherModel) FindWithCtx(ctx context.Context, query interface{}, opts ...*options.FindOneOptions) error {
	return FindOneWithCtx(ctx, m, query, opts...)
}

func (m *AnotherModel) FindByObjectID(id interface{}, opts ...*options.FindOneOptions) error {
	return FindByObjectID(m, id, opts...)
}

func (m *AnotherModel) FindByObjectIDWithCtx(ctx context.Context, id interface{}, opts ...*options.FindOneOptions) error {
	return FindByObjectIDWithCtx(ctx, m, id, opts...)
}

func (m *AnotherModel) Create(opts ...*options.InsertOneOptions) error {
	return InsertOne(m, opts...)
}

func (m *AnotherModel) CreateWithCtx(ctx context.Context, opts ...*options.InsertOneOptions) error {
	return InsertOneWithCtx(ctx, m, opts...)
}

func (m *AnotherModel) Update(opts ...*options.UpdateOptions) error {
	return Update(m, opts...)
}

func (m *AnotherModel) UpdateWithCtx(ctx context.Context, opts ...*options.UpdateOptions) error {
	return UpdateWithCtx(ctx, m, opts...)
}

func (m *AnotherModel) Delete(opts ...*options.DeleteOptions) error {
	return Delete(m, opts...)
}

func (m *AnotherModel) DeleteWithCtx(ctx context.Context, opts ...*options.DeleteOptions) error {
	return DeleteWithCtx(ctx, m, opts...)
}

func (m *Model) AggregateFirst(pipeline interface{}, opts ...*options.AggregateOptions) (bool, error) {
	return AggregateFirst(m, pipeline, opts...)
}

func (m *Model) Find(query interface{}, opts ...*options.FindOneOptions) error {
	return FindOne(m, query, opts...)
}

func (m *Model) FindWithCtx(ctx context.Context, query interface{}, opts ...*options.FindOneOptions) error {
	return FindOneWithCtx(ctx, m, query, opts...)
}

func (m *Model) FindByObjectID(id interface{}, opts ...*options.FindOneOptions) error {
	return FindByObjectID(m, id, opts...)
}

func (m *Model) FindByObjectIDWithCtx(ctx context.Context, id interface{}, opts ...*options.FindOneOptions) error {
	return FindByObjectIDWithCtx(ctx, m, id, opts...)
}

func (m *Model) Create(opts ...*options.InsertOneOptions) error {
	return InsertOne(m, opts...)
}

func (m *Model) CreateWithCtx(ctx context.Context, opts ...*options.InsertOneOptions) error {
	return InsertOneWithCtx(ctx, m, opts...)
}

func (m *Model) Update(opts ...*options.UpdateOptions) error {
	return Update(m, opts...)
}

func (m *Model) UpdateWithCtx(ctx context.Context, opts ...*options.UpdateOptions) error {
	return UpdateWithCtx(ctx, m, opts...)
}

func (m *Model) Delete(opts ...*options.DeleteOptions) error {
	return Delete(m, opts...)
}

func (m *Model) DeleteWithCtx(ctx context.Context, opts ...*options.DeleteOptions) error {
	return DeleteWithCtx(ctx, m, opts...)
}
