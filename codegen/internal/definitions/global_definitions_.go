package definitions

import (
	"context"
	"errors"
	"reflect"
	"time"

	"github.com/jonoans/mongo-gen/codegen"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type ModelInterface interface {
	CollectionName() string

	// Field Information
	GetID() any
	SetID(id any)

	// Hooks
	Queried() error
	Creating() error
	Created() error
	Saving() error
	Saved() error
	Updating() error
	Updated() error
	Deleting() error
	Deleted() error
}

// Available query methods
type ModelQueryMethods interface {
	AggregateFirst(pipeline any, opts ...options.Lister[options.AggregateOptions]) (bool, error)
	AggregateFirstWithCtx(ctx context.Context, pipeline any, opts ...options.Lister[options.AggregateOptions]) (bool, error)
	Find(any, ...options.Lister[options.FindOneOptions]) error
	FindWithCtx(context.Context, any, ...options.Lister[options.FindOneOptions]) error
	FindByObjectID(any, ...options.Lister[options.FindOneOptions]) error
	FindByObjectIDWithCtx(context.Context, any, ...options.Lister[options.FindOneOptions]) error
	Create(...options.Lister[options.InsertOneOptions]) error
	CreateWithCtx(context.Context, ...options.Lister[options.InsertOneOptions]) error
	Update(...options.Lister[options.UpdateOneOptions]) error
	UpdateWithCtx(context.Context, ...options.Lister[options.UpdateOneOptions]) error
	Delete(...options.Lister[options.DeleteOneOptions]) error
	DeleteWithCtx(context.Context, ...options.Lister[options.DeleteOneOptions]) error
}

type Config struct {
	OperationTimeout time.Duration
	DatabaseName     string

	TxnSessionOptions *options.SessionOptionsBuilder
}

func Initialise(cfg Config, opts ...*options.ClientOptions) error {
	if err := checkConfig(&cfg); err != nil {
		return err
	}

	if defaultClt != nil {
		return errors.New("client is already initialised")
	}

	defaultCfg = cfg
	client, err := mongo.Connect(opts...)
	if err != nil {
		return err
	}

	defaultClt = &databaseClient{
		client:      client,
		collections: map[string]*mongo.Collection{},
	}
	defaultClt.init()
	return nil
}

func GetClient() (*mongo.Client, error) {
	clt, err := getDefaultClient()
	if err != nil {
		return nil, err
	}
	return clt.client, nil
}

func GetDatabase() (*mongo.Database, error) {
	clt, err := getDefaultClient()
	if err != nil {
		return nil, err
	}
	return clt.getDatabase(), nil
}

func GetCollection(collectionName string) (*mongo.Collection, error) {
	clt, err := getDefaultClient()
	if err != nil {
		return nil, err
	}
	return clt.getCollection(collectionName), nil
}

func Coll(model ModelInterface) *mongo.Collection {
	name, err := getCollectionName(model)
	if err != nil {
		panic(err)
	}
	coll, _ := GetCollection(name)
	return coll
}

// Section: Query Functions

func Aggregate(results any, pipeline any, opts ...options.Lister[options.AggregateOptions]) error {
	ctx, cancel := newCtx()
	defer cancel()
	return AggregateWithCtx(ctx, results, pipeline, opts...)
}

func AggregateFirst(model ModelInterface, pipeline any, opts ...options.Lister[options.AggregateOptions]) (bool, error) {
	ctx, cancel := newCtx()
	defer cancel()
	return AggregateFirstWithCtx(ctx, model, pipeline, opts...)
}

func Delete(model ModelInterface, opts ...options.Lister[options.DeleteOneOptions]) error {
	ctx, cancel := newCtx()
	defer cancel()
	return DeleteWithCtx(ctx, model, opts...)
}

func DeleteOne(model ModelInterface, query any, opts ...options.Lister[options.DeleteOneOptions]) (*mongo.DeleteResult, error) {
	ctx, cancel := newCtx()
	defer cancel()
	return DeleteOneWithCtx(ctx, model, query, opts...)
}

func DeleteMany(model ModelInterface, query any, opts ...options.Lister[options.DeleteManyOptions]) (*mongo.DeleteResult, error) {
	ctx, cancel := newCtx()
	defer cancel()
	return DeleteManyWithCtx(ctx, model, query, opts...)
}

func FindOne(model ModelInterface, query any, opts ...options.Lister[options.FindOneOptions]) error {
	ctx, cancel := newCtx()
	defer cancel()
	return FindOneWithCtx(ctx, model, query, opts...)
}

func FindMany(results any, query any, opts ...options.Lister[options.FindOptions]) error {
	ctx, cancel := newCtx()
	defer cancel()
	return FindManyWithCtx(ctx, results, query, opts...)
}

func FindByObjectID(model ModelInterface, id any, opts ...options.Lister[options.FindOneOptions]) error {
	ctx, cancel := newCtx()
	defer cancel()
	return FindByObjectIDWithCtx(ctx, model, id, opts...)
}

func FindByObjectIDs(results any, ids any, additionalPipeline ...any) error {
	ctx, cancel := newCtx()
	defer cancel()
	return FindByObjectIDsWithCtx(ctx, results, ids, additionalPipeline...)
}

func FindByObjectIDsWithCtx(ctx context.Context, results any, ids any, additionalPipeline ...any) error {
	pipeline := bson.A{
		bson.M{"$match": bson.M{"_id": bson.M{"$in": ids}}},
		bson.M{"$addFields": bson.M{"_codegen_sort_index": bson.M{"$indexOfArray": bson.A{ids, "$_id"}}}},
		bson.M{"$sort": bson.M{"_codegen_sort_index": 1}},
		bson.M{"$project": bson.M{"_codegen_sort_index": 0}},
	}
	pipeline = append(pipeline, additionalPipeline...)
	return AggregateWithCtx(ctx, results, pipeline)
}

func InsertOne(model ModelInterface, opts ...options.Lister[options.InsertOneOptions]) error {
	ctx, cancel := newCtx()
	defer cancel()
	return InsertOneWithCtx(ctx, model, opts...)
}

func Update(model ModelInterface, opts ...options.Lister[options.UpdateOneOptions]) error {
	ctx, cancel := newCtx()
	defer cancel()
	return UpdateWithCtx(ctx, model, opts...)
}

func UpdateOne(model ModelInterface, filter any, update any, opts ...options.Lister[options.UpdateOneOptions]) (*mongo.UpdateResult, error) {
	ctx, cancel := newCtx()
	defer cancel()
	return UpdateOneWithCtx(ctx, model, filter, update, opts...)
}

func UpdateMany(model ModelInterface, filter any, update any, opts ...options.Lister[options.UpdateManyOptions]) (*mongo.UpdateResult, error) {
	ctx, cancel := newCtx()
	defer cancel()
	return UpdateManyWithCtx(ctx, model, filter, update, opts...)
}

// Section: Context Functions

func AggregateWithCtx(ctx context.Context, results any, pipeline any, aggregateOpts ...options.Lister[options.AggregateOptions]) error {
	collectionName, err := getCollectionNameFromSlice(results)
	if err != nil {
		return err
	}

	collection, err := GetCollection(collectionName)
	if err != nil {
		return err
	}

	cur, err := collection.Aggregate(ctx, pipeline, aggregateOpts...)
	if cur != nil {
		defer cur.Close(ctx)
	}

	if err != nil {
		return err
	}

	if err := cur.All(ctx, results); err != nil {
		return err
	}

	if err := runFuncOnResultsSliceItems(results, callAfterQueryHooks); err != nil {
		return err
	}

	return nil
}

func AggregateFirstWithCtx(ctx context.Context, result ModelInterface, pipeline any, aggregateOpts ...options.Lister[options.AggregateOptions]) (bool, error) {
	collectionName, err := getCollectionName(result)
	if err != nil {
		return false, err
	}

	collection, err := GetCollection(collectionName)
	if err != nil {
		return false, err
	}

	cur, err := collection.Aggregate(ctx, pipeline, aggregateOpts...)
	if cur != nil {
		defer cur.Close(ctx)
	}

	if err != nil {
		return false, err
	}

	if cur.Next(ctx) {
		if err := cur.Decode(result); err != nil {
			return false, err
		}
		return true, callAfterQueryHooks(result)
	}

	return false, nil
}

func DeleteWithCtx(ctx context.Context, model ModelInterface, opts ...options.Lister[options.DeleteOneOptions]) error {
	if err := callBeforeDeleteHooks(model); err != nil {
		return err
	}

	if _, err := DeleteOneWithCtx(ctx, model, bson.M{"_id": model.GetID()}, opts...); err != nil {
		return err
	}

	return callAfterDeleteHooks(model)
}

func DeleteOneWithCtx(ctx context.Context, model ModelInterface, query any, opts ...options.Lister[options.DeleteOneOptions]) (*mongo.DeleteResult, error) {
	collectionName, err := getCollectionName(model)
	if err != nil {
		return nil, err
	}

	coll, err := GetCollection(collectionName)
	if err != nil {
		return nil, err
	}

	result, err := coll.DeleteOne(ctx, query, opts...)
	if err != nil {
		return result, err
	}

	return result, nil
}

func DeleteManyWithCtx(ctx context.Context, model ModelInterface, query any, opts ...options.Lister[options.DeleteManyOptions]) (*mongo.DeleteResult, error) {
	collectionName, err := getCollectionName(model)
	if err != nil {
		return nil, err
	}

	coll, err := GetCollection(collectionName)
	if err != nil {
		return nil, err
	}

	result, err := coll.DeleteMany(ctx, query, opts...)
	if err != nil {
		return result, err
	}

	return result, nil
}

func FindOneWithCtx(ctx context.Context, model ModelInterface, query any, opts ...options.Lister[options.FindOneOptions]) error {
	collectionName, err := getCollectionName(model)
	if err != nil {
		return err
	}

	coll, err := GetCollection(collectionName)
	if err != nil {
		return err
	}

	if err := coll.FindOne(ctx, query, opts...).Decode(model); err != nil {
		return err
	}

	return callAfterQueryHooks(model)
}

func FindManyWithCtx(ctx context.Context, results any, query any, opts ...options.Lister[options.FindOptions]) error {
	collectionName, err := getCollectionNameFromSlice(results)
	if err != nil {
		return err
	}

	coll, err := GetCollection(collectionName)
	if err != nil {
		return err
	}

	cur, err := coll.Find(ctx, query, opts...)
	if cur != nil {
		defer cur.Close(ctx)
	}

	if err != nil {
		return err
	}

	if err := cur.All(ctx, results); err != nil {
		return err
	}

	if err := runFuncOnResultsSliceItems(results, callAfterQueryHooks); err != nil {
		return err
	}

	return nil
}

func FindByObjectIDWithCtx(ctx context.Context, model ModelInterface, id any, opts ...options.Lister[options.FindOneOptions]) error {
	oid, err := assertObjectID(id)
	if err != nil {
		return err
	}
	return FindOneWithCtx(ctx, model, bson.M{"_id": oid}, opts...)
}

func InsertOneWithCtx(ctx context.Context, model ModelInterface, opts ...options.Lister[options.InsertOneOptions]) error {
	collectionName, err := getCollectionName(model)
	if err != nil {
		return err
	}

	if err := callBeforeCreateHooks(model); err != nil {
		return err
	}

	coll, err := GetCollection(collectionName)
	if err != nil {
		return err
	}

	result, err := coll.InsertOne(ctx, model, opts...)
	if err != nil {
		return err
	}

	return callAfterCreateHooks(model, result)
}

func UpdateWithCtx(ctx context.Context, model ModelInterface, opts ...options.Lister[options.UpdateOneOptions]) error {
	collectionName, err := getCollectionName(model)
	if err != nil {
		return err
	}

	if err := callBeforeUpdateHooks(model); err != nil {
		return err
	}

	coll, err := GetCollection(collectionName)
	if err != nil {
		return err
	}

	if _, err = coll.UpdateByID(ctx, model.GetID(), bson.M{"$set": model}, opts...); err != nil {
		return err
	}

	return callAfterUpdateHooks(model)
}

func UpdateOneWithCtx(ctx context.Context, model ModelInterface, filter any, update any, opts ...options.Lister[options.UpdateOneOptions]) (*mongo.UpdateResult, error) {
	collectionName, err := getCollectionName(model)
	if err != nil {
		return nil, err
	}

	coll, err := GetCollection(collectionName)
	if err != nil {
		return nil, err
	}

	result, err := coll.UpdateOne(ctx, filter, update, opts...)
	if err != nil {
		return result, err
	}

	return result, nil
}

func UpdateManyWithCtx(ctx context.Context, model ModelInterface, filter any, update any, opts ...options.Lister[options.UpdateManyOptions]) (*mongo.UpdateResult, error) {
	collectionName, err := getCollectionName(model)
	if err != nil {
		return nil, err
	}

	coll, err := GetCollection(collectionName)
	if err != nil {
		return nil, err
	}

	result, err := coll.UpdateMany(ctx, filter, update, opts...)
	if err != nil {
		return result, err
	}

	return result, nil
}

func Transaction(fn codegen.TransactionFunc) error {
	ctx, cancel := newCtx()
	defer cancel()
	return TransactionWithCtx(ctx, fn)
}

func TransactionWithCtx(ctx context.Context, fn codegen.TransactionFunc) error {
	return TransactionWithCtxOptions(ctx, fn, defaultCfg.TxnSessionOptions)
}

func TransactionWithOptions(fn codegen.TransactionFunc, opts *options.SessionOptionsBuilder) error {
	ctx, cancel := newCtx()
	defer cancel()
	return TransactionWithCtxOptions(ctx, fn, opts)
}

func TransactionWithCtxOptions(ctx context.Context, fn codegen.TransactionFunc, opts *options.SessionOptionsBuilder) error {
	client, err := GetClient()
	if err != nil {
		return err
	}

	return client.UseSessionWithOptions(ctx, opts, func(ctx context.Context) error {
		sess := mongo.SessionFromContext(ctx)
		if err := sess.StartTransaction(); err != nil {
			return err
		}
		return fn(ctx)
	})
}

func Close() {
	if defaultClt != nil {
		ctx, cancel := newCtx()
		defer cancel()
		_ = defaultClt.client.Disconnect(ctx)
		defaultClt = nil
	}
}

// Section: Private Functions

type databaseClient struct {
	client      *mongo.Client
	database    *mongo.Database
	collections map[string]*mongo.Collection
}

func (c *databaseClient) init() {
	c.collections = make(map[string]*mongo.Collection)
}

func (c *databaseClient) getDatabase() *mongo.Database {
	if c.database == nil {
		c.database = c.client.Database(defaultCfg.DatabaseName)
	}
	return c.database
}

func (c *databaseClient) getCollection(name string) *mongo.Collection {
	if coll, ok := c.collections[name]; ok {
		return coll
	}
	c.collections[name] = c.getDatabase().Collection(name)
	return c.collections[name]
}

var (
	defaultClt *databaseClient
	defaultCfg Config
)

func Ctx() context.Context {
	ctx, _ := newCtx()
	return ctx
}

func newCtx() (context.Context, func()) {
	// Can't cancel context
	return context.WithTimeout(context.Background(), defaultCfg.OperationTimeout)
}

func getDefaultClient() (*databaseClient, error) {
	if defaultClt == nil {
		return nil, errors.New("client is not initialised, please call the Initialise method first!")
	}
	return defaultClt, nil
}

func assertObjectID(id any) (bson.ObjectID, error) {
	switch v := id.(type) {
	case bson.ObjectID:
		return v, nil
	case *bson.ObjectID:
		return *v, nil
	case string:
		return bson.ObjectIDFromHex(v)
	default:
		return bson.NilObjectID, errors.New("invalid object id")
	}
}

func checkConfig(cfg *Config) error {
	if cfg.DatabaseName == "" {
		return errors.New("database name is empty")
	}

	// Fill default
	if cfg.OperationTimeout == 0 {
		cfg.OperationTimeout = time.Second * 15
	}

	if cfg.TxnSessionOptions == nil {
		cfg.TxnSessionOptions = options.Session()
	}

	return nil
}

func getCollectionName(model ModelInterface) (string, error) {
	name := model.CollectionName()
	if name == "" {
		return "", errors.New("collection name is empty")
	}
	return name, nil
}

func getCollectNameFromInterface(model any) (string, error) {
	if v, ok := model.(ModelInterface); ok {
		return getCollectionName(v)
	}
	return "", errors.New("model is not a ModelInterface")
}

func getCollectionNameFromSlice(results any) (string, error) {
	resultsType := reflect.TypeOf(results)
	if resultsType.Kind() != reflect.Ptr {
		return "", errors.New("results is not a pointer")
	}

	resultsType = reflect.Indirect(reflect.ValueOf(results)).Type()
	if resultsType.Kind() != reflect.Slice {
		return "", errors.New("results is not a pointer to a slice")
	}

	elemValue := reflect.New(resultsType.Elem()).Interface()
	return getCollectNameFromInterface(elemValue)
}

// Only use when sure results is slice of ModelInterface
func runFuncOnResultsSliceItems(results any, callback func(model ModelInterface) error) error {
	resultsPtr := reflect.ValueOf(results)
	resultsSlice := reflect.Indirect(resultsPtr)
	resultsSliceLen := resultsSlice.Len()
	for i := range resultsSliceLen {
		item := resultsSlice.Index(i).Addr().Interface()
		if err := callback(item.(ModelInterface)); err != nil {
			return err
		}
	}
	return nil
}

// Section: Hook Helpers

func callAfterQueryHooks(model ModelInterface) error {
	if err := model.Queried(); err != nil {
		return err
	}

	return nil
}

func callBeforeCreateHooks(model ModelInterface) error {
	if err := model.Creating(); err != nil {
		return err
	}

	if err := model.Saving(); err != nil {
		return err
	}

	return nil
}

func callAfterCreateHooks(model ModelInterface, result *mongo.InsertOneResult) error {
	model.SetID(result.InsertedID)

	if err := model.Created(); err != nil {
		return err
	}

	if err := model.Saved(); err != nil {
		return err
	}

	return nil
}

func callBeforeUpdateHooks(model ModelInterface) error {
	if err := model.Updating(); err != nil {
		return err
	}

	if err := model.Saving(); err != nil {
		return err
	}

	return nil
}

func callAfterUpdateHooks(model ModelInterface) error {
	if err := model.Updated(); err != nil {
		return err
	}

	if err := model.Saved(); err != nil {
		return err
	}

	return nil
}

func callBeforeDeleteHooks(model ModelInterface) error {
	if err := model.Deleting(); err != nil {
		return err
	}

	return nil
}

func callAfterDeleteHooks(model ModelInterface) error {
	if err := model.Deleted(); err != nil {
		return err
	}

	return nil
}
