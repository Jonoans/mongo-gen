package definitions

import (
	"context"
	"errors"
	"reflect"
	"time"

	"github.com/jonoans/mongo-gen/codegen"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ModelInterface interface {
	CollectionName() string

	GetID() interface{}
	SetID(id interface{})

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

type Config struct {
	OperationTimeout time.Duration
	DatabaseName     string

	TxnSessionOptions *options.SessionOptions
}

func Initialise(cfg Config, opts ...*options.ClientOptions) error {
	if err := checkConfig(&cfg); err != nil {
		return err
	}

	defaultCfg = cfg
	client, err := mongo.NewClient(opts...)
	if err != nil {
		return err
	}

	if defaultClt != nil {
		return errors.New("client is already initialised")
	}

	if err := client.Connect(newCtx()); err != nil {
		return err
	}

	defaultClt = &databaseClient{
		client:      client,
		collections: map[string]*mongo.Collection{},
	}
	return nil
}

func GetClient() (*mongo.Client, error) {
	if defaultClt == nil {
		return nil, errors.New("client is not initialised, please call the Initialise method first!")
	}
	return defaultClt.client, nil
}

func GetDatabase() (*mongo.Database, error) {
	if defaultClt.database == nil {
		client, err := GetClient()
		if err != nil {
			return nil, err
		}

		defaultClt.database = client.Database(defaultCfg.DatabaseName)
	}

	return defaultClt.database, nil
}

func GetCollection(collectionName string) (*mongo.Collection, error) {
	if c, ok := defaultClt.collections[collectionName]; ok {
		return c, nil
	}

	db, err := GetDatabase()
	if err != nil {
		return nil, err
	}

	defaultClt.collections[collectionName] = db.Collection(collectionName)
	return defaultClt.collections[collectionName], nil
}

// Section: Query Functions

func Aggregate(results interface{}, pipeline interface{}, opts ...*options.AggregateOptions) error {
	return AggregateWithCtx(newCtx(), results, pipeline, opts...)
}

func DeleteOne(model ModelInterface, opts ...*options.DeleteOptions) error {
	return DeleteOneWithCtx(newCtx(), model, opts...)
}

func FindOne(model ModelInterface, query interface{}, opts ...*options.FindOneOptions) error {
	return FindOneWithCtx(newCtx(), model, query, opts...)
}

func FindByUUID(model ModelInterface, uuidFieldName, uuid string, opts ...*options.FindOneOptions) error {
	return FindOneWithCtx(newCtx(), model, bson.M{uuidFieldName: uuid}, opts...)
}

func FindByObjectID(model ModelInterface, id interface{}, opts ...*options.FindOneOptions) error {
	var oid primitive.ObjectID

	switch v := id.(type) {
	case primitive.ObjectID:
		oid = v
	case *primitive.ObjectID:
		oid = *v
	case string:
		id, err := primitive.ObjectIDFromHex(v)
		if err != nil {
			return err
		}
		oid = id
	default:
		return errors.New("invalid id")
	}

	return FindOneWithCtx(newCtx(), model, bson.M{"_id": oid}, opts...)
}

func FindByObjectIDs(results interface{}, ids interface{}) error {
	pipeline := bson.A{
		bson.M{"$match": bson.M{"_id": bson.M{"$in": ids}}},
		bson.M{"$addFields": bson.M{"_codegen_sort_index": bson.M{"$indexOfArray": bson.A{ids, "$_id"}}}},
		bson.M{"$sort": bson.M{"_codegen_sort_index": 1}},
		bson.M{"$project": bson.M{"_codegen_sort_index": 0}},
	}
	return Aggregate(results, pipeline)
}

func InsertOne(model ModelInterface, opts ...*options.InsertOneOptions) error {
	return InsertOneWithCtx(newCtx(), model, opts...)
}

func Update(model ModelInterface, opts ...*options.UpdateOptions) error {
	return UpdateWithCtx(newCtx(), model, opts...)
}

// Section: Context Functions

func AggregateWithCtx(ctx context.Context, results interface{}, pipeline interface{}, aggregateOpts ...*options.AggregateOptions) error {
	collectionName, err := getCollectionNameFromSlice(results)
	if err != nil {
		return err
	}

	collection, err := GetCollection(collectionName)
	if err != nil {
		return err
	}

	cur, err := collection.Aggregate(ctx, pipeline, aggregateOpts...)
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

func AggregateFirstWithCtx(ctx context.Context, result ModelInterface, pipeline interface{}, aggregateOpts ...*options.AggregateOptions) (bool, error) {
	collectionName, err := getCollectionName(result)
	if err != nil {
		return false, err
	}

	collection, err := GetCollection(collectionName)
	if err != nil {
		return false, err
	}

	cur, err := collection.Aggregate(ctx, pipeline, aggregateOpts...)
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

func DeleteOneWithCtx(ctx context.Context, model ModelInterface, opts ...*options.DeleteOptions) error {
	collectionName, err := getCollectionName(model)
	if err != nil {
		return err
	}

	if err := callBeforeDeleteHooks(model); err != nil {
		return err
	}

	coll, err := GetCollection(collectionName)
	if err != nil {
		return err
	}

	if _, err := coll.DeleteOne(ctx, bson.M{"_id": model.GetID()}, opts...); err != nil {
		return err
	}

	return callAfterDeleteHooks(model)
}

func FindOneWithCtx(ctx context.Context, model ModelInterface, query interface{}, opts ...*options.FindOneOptions) error {
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

func InsertOneWithCtx(ctx context.Context, model ModelInterface, opts ...*options.InsertOneOptions) error {
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

func UpdateWithCtx(ctx context.Context, model ModelInterface, opts ...*options.UpdateOptions) error {
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

func Transaction(fn codegen.TransactionFunc) error {
	return TransactionWithOptions(fn, defaultCfg.TxnSessionOptions)
}

func TransactionWithOptions(fn codegen.TransactionFunc, opts *options.SessionOptions) error {
	client, err := GetClient()
	if err != nil {
		return err
	}

	return client.UseSessionWithOptions(newCtx(), opts, func(ctx mongo.SessionContext) error {
		ctx.StartTransaction()
		return fn(ctx)
	})
}

func Close() {
	if defaultClt != nil {
		defaultClt.client.Disconnect(newCtx())
		defaultClt = nil
	}
}

// Section: Private Functions

type databaseClient struct {
	client      *mongo.Client
	database    *mongo.Database
	collections map[string]*mongo.Collection
}

var (
	defaultClt *databaseClient
	defaultCfg Config
)

func newCtx() context.Context {
	// Can't cancel context
	ctx, _ := context.WithTimeout(context.Background(), defaultCfg.OperationTimeout)
	return ctx
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

func getCollectNameFromInterface(model interface{}) (string, error) {
	if v, ok := model.(ModelInterface); ok {
		return getCollectionName(v)
	}
	return "", errors.New("model is not a ModelInterface")
}

func getCollectionNameFromSlice(results interface{}) (string, error) {
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
func runFuncOnResultsSliceItems(results interface{}, callback func(model ModelInterface) error) error {
	resultsPtr := reflect.ValueOf(results)
	resultsSlice := reflect.Indirect(resultsPtr)
	resultsSliceLen := resultsSlice.Len()
	for i := 0; i < resultsSliceLen; i++ {
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
