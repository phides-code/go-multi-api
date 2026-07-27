// Unit tests for the car DynamoDB repository using a mocked DynamoDB client.
package car_test

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	awsdynamodb "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	"github.com/phides-code/go-multi-api/internal/car"
	"github.com/phides-code/go-multi-api/internal/domain"
	"github.com/phides-code/go-multi-api/internal/testutil"
)

type mockDynamoClient struct {
	getItemFn    func(ctx context.Context, params *awsdynamodb.GetItemInput, optFns ...func(*awsdynamodb.Options)) (*awsdynamodb.GetItemOutput, error)
	deleteItemFn func(ctx context.Context, params *awsdynamodb.DeleteItemInput, optFns ...func(*awsdynamodb.Options)) (*awsdynamodb.DeleteItemOutput, error)
	updateItemFn func(ctx context.Context, params *awsdynamodb.UpdateItemInput, optFns ...func(*awsdynamodb.Options)) (*awsdynamodb.UpdateItemOutput, error)
	putItemFn    func(ctx context.Context, params *awsdynamodb.PutItemInput, optFns ...func(*awsdynamodb.Options)) (*awsdynamodb.PutItemOutput, error)
	scanFn       func(ctx context.Context, params *awsdynamodb.ScanInput, optFns ...func(*awsdynamodb.Options)) (*awsdynamodb.ScanOutput, error)
}

func (m *mockDynamoClient) GetItem(ctx context.Context, params *awsdynamodb.GetItemInput, optFns ...func(*awsdynamodb.Options)) (*awsdynamodb.GetItemOutput, error) {
	return m.getItemFn(ctx, params, optFns...)
}

func (m *mockDynamoClient) PutItem(ctx context.Context, params *awsdynamodb.PutItemInput, optFns ...func(*awsdynamodb.Options)) (*awsdynamodb.PutItemOutput, error) {
	return m.putItemFn(ctx, params, optFns...)
}

func (m *mockDynamoClient) Scan(ctx context.Context, params *awsdynamodb.ScanInput, optFns ...func(*awsdynamodb.Options)) (*awsdynamodb.ScanOutput, error) {
	return m.scanFn(ctx, params, optFns...)
}

func (m *mockDynamoClient) UpdateItem(ctx context.Context, params *awsdynamodb.UpdateItemInput, optFns ...func(*awsdynamodb.Options)) (*awsdynamodb.UpdateItemOutput, error) {
	return m.updateItemFn(ctx, params, optFns...)
}

func (m *mockDynamoClient) DeleteItem(ctx context.Context, params *awsdynamodb.DeleteItemInput, optFns ...func(*awsdynamodb.Options)) (*awsdynamodb.DeleteItemOutput, error) {
	return m.deleteItemFn(ctx, params, optFns...)
}

func scanItems(t *testing.T, cars []car.Car) []map[string]types.AttributeValue {
	t.Helper()
	items := make([]map[string]types.AttributeValue, len(cars))
	for i, c := range cars {
		item, err := attributevalue.MarshalMap(c)
		if err != nil {
			t.Fatal(err)
		}
		items[i] = item
	}
	return items
}

func TestCarRepositoryGetByID(t *testing.T) {
	t.Parallel()

	validId, validCar, item := storedCarFixture(t)
	errSDK := errors.New("dynamo unavailable")
	tests := []struct {
		name     string
		setupMock func(t *testing.T) *mockDynamoClient
		wantCar  car.Car
		wantErr  error
	}{
		{
			name: "found",
			setupMock: func(_ *testing.T) *mockDynamoClient {
				return &mockDynamoClient{
					getItemFn: func(_ context.Context, _ *awsdynamodb.GetItemInput, _ ...func(*awsdynamodb.Options)) (*awsdynamodb.GetItemOutput, error) {
						return &awsdynamodb.GetItemOutput{Item: item}, nil
					},
				}
			},
			wantCar: validCar,
			wantErr: nil,
		},
		{
			name: "not found",
			setupMock: func(_ *testing.T) *mockDynamoClient {
				return &mockDynamoClient{
					getItemFn: func(_ context.Context, _ *awsdynamodb.GetItemInput, _ ...func(*awsdynamodb.Options)) (*awsdynamodb.GetItemOutput, error) {
						return &awsdynamodb.GetItemOutput{Item: nil}, nil
					},
				}
			},
			wantCar: car.Car{},
			wantErr: domain.ErrNotFound,
		},
		{
			name: "sdk error",
			setupMock: func(_ *testing.T) *mockDynamoClient {
				return &mockDynamoClient{
					getItemFn: func(_ context.Context, _ *awsdynamodb.GetItemInput, _ ...func(*awsdynamodb.Options)) (*awsdynamodb.GetItemOutput, error) {
						return nil, errSDK
					},
				}
			},
			wantCar: car.Car{},
			wantErr: errSDK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := car.NewRepository(tt.setupMock(t))
			got, err := repo.GetByID(context.Background(), validId)

			assertCarRepoResult(t, "GetByID", got, err, tt.wantCar, tt.wantErr)
		})
	}
}

func TestCarRepositoryDelete(t *testing.T) {
	t.Parallel()

	validId, validCar, item := storedCarFixture(t)
	errSDK := errors.New("dynamo unavailable")
	tests := []struct {
		name      string
		setupMock func(t *testing.T) *mockDynamoClient
		wantCar   car.Car
		wantErr   error
	}{
		{
			name: "success",
			setupMock: func(_ *testing.T) *mockDynamoClient {
				return &mockDynamoClient{
					deleteItemFn: func(_ context.Context, _ *awsdynamodb.DeleteItemInput, _ ...func(*awsdynamodb.Options)) (*awsdynamodb.DeleteItemOutput, error) {
						return &awsdynamodb.DeleteItemOutput{Attributes: item}, nil
					},
				}
			},
			wantCar: validCar,
			wantErr: nil,
		},
		{
			name: "not found",
			setupMock: func(_ *testing.T) *mockDynamoClient {
				return &mockDynamoClient{
					deleteItemFn: func(_ context.Context, _ *awsdynamodb.DeleteItemInput, _ ...func(*awsdynamodb.Options)) (*awsdynamodb.DeleteItemOutput, error) {
						return &awsdynamodb.DeleteItemOutput{Attributes: nil}, nil
					},
				}
			},
			wantCar: car.Car{},
			wantErr: domain.ErrNotFound,
		},
		{
			name: "sdk error",
			setupMock: func(_ *testing.T) *mockDynamoClient {
				return &mockDynamoClient{
					deleteItemFn: func(_ context.Context, _ *awsdynamodb.DeleteItemInput, _ ...func(*awsdynamodb.Options)) (*awsdynamodb.DeleteItemOutput, error) {
						return nil, errSDK
					},
				}
			},
			wantCar: car.Car{},
			wantErr: errSDK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := car.NewRepository(tt.setupMock(t))
			got, err := repo.Delete(context.Background(), validId)

			assertCarRepoResult(t, "Delete", got, err, tt.wantCar, tt.wantErr)
		})
	}
}

func TestCarRepositoryUpdate(t *testing.T) {
	t.Parallel()

	updatedCar := car.Car{ID: uuid.NewString(), Model: "updated", CreatedOn: 12345}
	errSDK := errors.New("dynamo unavailable")

	item, err := attributevalue.MarshalMap(updatedCar)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name      string
		setupMock func(t *testing.T) *mockDynamoClient
		wantCar   car.Car
		wantErr   error
	}{
		{
			name: "success",
			setupMock: func(t *testing.T) *mockDynamoClient {
				return &mockDynamoClient{
					updateItemFn: func(_ context.Context, params *awsdynamodb.UpdateItemInput, _ ...func(*awsdynamodb.Options)) (*awsdynamodb.UpdateItemOutput, error) {
						testutil.AssertUpdateSets(t, params, map[string]any{
							"model": updatedCar.Model,
						})
						return &awsdynamodb.UpdateItemOutput{Attributes: item}, nil
					},
				}
			},
			wantCar: updatedCar,
			wantErr: nil,
		},
		{
			name: "not found",
			setupMock: func(_ *testing.T) *mockDynamoClient {
				return &mockDynamoClient{
					updateItemFn: func(_ context.Context, _ *awsdynamodb.UpdateItemInput, _ ...func(*awsdynamodb.Options)) (*awsdynamodb.UpdateItemOutput, error) {
						return nil, &types.ConditionalCheckFailedException{}
					},
				}
			},
			wantCar: car.Car{},
			wantErr: domain.ErrNotFound,
		},
		{
			name: "sdk error",
			setupMock: func(_ *testing.T) *mockDynamoClient {
				return &mockDynamoClient{
					updateItemFn: func(_ context.Context, _ *awsdynamodb.UpdateItemInput, _ ...func(*awsdynamodb.Options)) (*awsdynamodb.UpdateItemOutput, error) {
						return nil, errSDK
					},
				}
			},
			wantCar: car.Car{},
			wantErr: errSDK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := car.NewRepository(tt.setupMock(t))
			got, err := repo.Update(context.Background(), updatedCar)

			assertCarRepoResult(t, "Update", got, err, tt.wantCar, tt.wantErr)
		})
	}
}

func TestCarRepositoryCreate(t *testing.T) {
	t.Parallel()

	want := car.Car{ID: uuid.NewString(), Model: "new", CreatedOn: 12345}
	errSDK := errors.New("dynamo unavailable")

	tests := []struct {
		name      string
		setupMock func(t *testing.T) *mockDynamoClient
		wantCar   car.Car
		wantErr   error
	}{
		{
			name: "success",
			setupMock: func(t *testing.T) *mockDynamoClient {
				return &mockDynamoClient{
					putItemFn: func(_ context.Context, params *awsdynamodb.PutItemInput, _ ...func(*awsdynamodb.Options)) (*awsdynamodb.PutItemOutput, error) {
						assertCarPutItem(t, params, want)
						return &awsdynamodb.PutItemOutput{}, nil
					},
				}
			},
			wantCar: want,
			wantErr: nil,
		},
		{
			name: "duplicate id",
			setupMock: func(_ *testing.T) *mockDynamoClient {
				return &mockDynamoClient{
					putItemFn: func(_ context.Context, _ *awsdynamodb.PutItemInput, _ ...func(*awsdynamodb.Options)) (*awsdynamodb.PutItemOutput, error) {
						return nil, &types.ConditionalCheckFailedException{}
					},
				}
			},
			wantCar: car.Car{},
			wantErr: domain.ErrAlreadyExists,
		},
		{
			name: "sdk error",
			setupMock: func(_ *testing.T) *mockDynamoClient {
				return &mockDynamoClient{
					putItemFn: func(_ context.Context, _ *awsdynamodb.PutItemInput, _ ...func(*awsdynamodb.Options)) (*awsdynamodb.PutItemOutput, error) {
						return nil, errSDK
					},
				}
			},
			wantCar: car.Car{},
			wantErr: errSDK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := car.NewRepository(tt.setupMock(t))
			got, err := repo.Create(context.Background(), want)

			assertCarRepoResult(t, "Create", got, err, tt.wantCar, tt.wantErr)
		})
	}
}

func TestCarRepositoryList(t *testing.T) {
	t.Parallel()

	c1, c2, c3 := testutil.ListCars(true)
	wantItems := []car.Car{c1, c2}
	page2 := []car.Car{c3}
	scanOutputItems := scanItems(t, wantItems)
	page2ScanItems := scanItems(t, page2)

	tests := []struct {
		name      string
		setupMock func(t *testing.T) *mockDynamoClient
		wantItems []car.Car
		wantErr   bool
	}{
		{
			name: "returns items",
			setupMock: func(_ *testing.T) *mockDynamoClient {
				return &mockDynamoClient{
					scanFn: func(_ context.Context, params *awsdynamodb.ScanInput, _ ...func(*awsdynamodb.Options)) (*awsdynamodb.ScanOutput, error) {
						if params.Limit != nil {
							t.Errorf("Limit = %v, want nil", params.Limit)
						}
						return &awsdynamodb.ScanOutput{Items: scanOutputItems}, nil
					},
				}
			},
			wantItems: wantItems,
		},
		{
			name: "empty",
			setupMock: func(_ *testing.T) *mockDynamoClient {
				return &mockDynamoClient{
					scanFn: func(_ context.Context, _ *awsdynamodb.ScanInput, _ ...func(*awsdynamodb.Options)) (*awsdynamodb.ScanOutput, error) {
						return &awsdynamodb.ScanOutput{Items: nil}, nil
					},
				}
			},
			wantItems: []car.Car{},
		},
		{
			name: "scans all pages",
			setupMock: func(_ *testing.T) *mockDynamoClient {
				calls := 0
				return &mockDynamoClient{
					scanFn: func(_ context.Context, params *awsdynamodb.ScanInput, _ ...func(*awsdynamodb.Options)) (*awsdynamodb.ScanOutput, error) {
						calls++
						switch calls {
						case 1:
							if params.ExclusiveStartKey != nil {
								t.Fatal("expected first scan without ExclusiveStartKey")
							}
							return &awsdynamodb.ScanOutput{
								Items: scanOutputItems,
								LastEvaluatedKey: map[string]types.AttributeValue{
									"id": &types.AttributeValueMemberS{Value: c2.ID},
								},
							}, nil
						case 2:
							if params.ExclusiveStartKey == nil {
								t.Fatal("expected second scan with ExclusiveStartKey")
							}
							return &awsdynamodb.ScanOutput{Items: page2ScanItems}, nil
						default:
							t.Fatal("unexpected extra scan")
							return nil, nil
						}
					},
				}
			},
			wantItems: append(wantItems, page2...),
		},
		{
			name: "sdk error",
			setupMock: func(_ *testing.T) *mockDynamoClient {
				return &mockDynamoClient{
					scanFn: func(_ context.Context, _ *awsdynamodb.ScanInput, _ ...func(*awsdynamodb.Options)) (*awsdynamodb.ScanOutput, error) {
						return nil, errors.New("dynamo unavailable")
					},
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := car.NewRepository(tt.setupMock(t))
			items, err := repo.List(context.Background())

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("List: %v", err)
			}

			if len(items) != len(tt.wantItems) {
				t.Fatalf("len(items) = %d, want %d", len(items), len(tt.wantItems))
			}

			for i := range tt.wantItems {
				if items[i] != tt.wantItems[i] {
					t.Fatalf("items[%d] = %+v, want %+v", i, items[i], tt.wantItems[i])
				}
			}
		})
	}
}
