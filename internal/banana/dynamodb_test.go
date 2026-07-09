// Unit tests for the banana DynamoDB repository using a mocked DynamoDB client.
package banana_test

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	awsdynamodb "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	"github.com/phides-code/go-multi-api/internal/banana"
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

func scanItems(t *testing.T, bananas []banana.Banana) []map[string]types.AttributeValue {
	t.Helper()
	items := make([]map[string]types.AttributeValue, len(bananas))
	for i, b := range bananas {
		item, err := attributevalue.MarshalMap(b)
		if err != nil {
			t.Fatal(err)
		}
		items[i] = item
	}
	return items
}

func TestBananaRepositoryGetByID(t *testing.T) {
	t.Parallel()

	validId, validBanana, item := storedBananaFixture(t)
	errSDK := errors.New("dynamo unavailable")
	tests := []struct {
		name       string
		setupMock  func(t *testing.T) *mockDynamoClient
		wantBanana banana.Banana
		wantErr    error
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
			wantBanana: validBanana,
			wantErr:    nil,
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
			wantBanana: banana.Banana{},
			wantErr:    domain.ErrNotFound,
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
			wantBanana: banana.Banana{},
			wantErr:    errSDK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := banana.NewRepository(tt.setupMock(t))
			got, err := repo.GetByID(context.Background(), validId)

			assertBananaRepoResult(t, "GetByID", got, err, tt.wantBanana, tt.wantErr)
		})
	}
}

func TestBananaRepositoryDelete(t *testing.T) {
	t.Parallel()

	validId, validBanana, item := storedBananaFixture(t)
	errSDK := errors.New("dynamo unavailable")
	tests := []struct {
		name       string
		setupMock  func(t *testing.T) *mockDynamoClient
		wantBanana banana.Banana
		wantErr    error
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
			wantBanana: validBanana,
			wantErr:    nil,
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
			wantBanana: banana.Banana{},
			wantErr:    domain.ErrNotFound,
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
			wantBanana: banana.Banana{},
			wantErr:    errSDK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := banana.NewRepository(tt.setupMock(t))
			got, err := repo.Delete(context.Background(), validId)

			assertBananaRepoResult(t, "Delete", got, err, tt.wantBanana, tt.wantErr)
		})
	}
}

func TestBananaRepositoryUpdate(t *testing.T) {
	t.Parallel()

	updatedBanana := banana.Banana{ID: uuid.NewString(), Content: "updated", CreatedOn: 12345}
	errSDK := errors.New("dynamo unavailable")

	item, err := attributevalue.MarshalMap(updatedBanana)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name       string
		setupMock  func(t *testing.T) *mockDynamoClient
		wantBanana banana.Banana
		wantErr    error
	}{
		{
			name: "success",
			setupMock: func(t *testing.T) *mockDynamoClient {
				return &mockDynamoClient{
					updateItemFn: func(_ context.Context, params *awsdynamodb.UpdateItemInput, _ ...func(*awsdynamodb.Options)) (*awsdynamodb.UpdateItemOutput, error) {
						testutil.AssertUpdateSets(t, params, map[string]string{
							"content": updatedBanana.Content,
						})
						return &awsdynamodb.UpdateItemOutput{Attributes: item}, nil
					},
				}
			},
			wantBanana: updatedBanana,
			wantErr:    nil,
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
			wantBanana: banana.Banana{},
			wantErr:    domain.ErrNotFound,
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
			wantBanana: banana.Banana{},
			wantErr:    errSDK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := banana.NewRepository(tt.setupMock(t))
			got, err := repo.Update(context.Background(), updatedBanana)

			assertBananaRepoResult(t, "Update", got, err, tt.wantBanana, tt.wantErr)
		})
	}
}

func TestBananaRepositoryCreate(t *testing.T) {
	t.Parallel()

	want := banana.Banana{ID: uuid.NewString(), Content: "new", CreatedOn: 12345}
	errSDK := errors.New("dynamo unavailable")

	tests := []struct {
		name       string
		setupMock  func(t *testing.T) *mockDynamoClient
		wantBanana banana.Banana
		wantErr    error
	}{
		{
			name: "success",
			setupMock: func(t *testing.T) *mockDynamoClient {
				return &mockDynamoClient{
					putItemFn: func(_ context.Context, params *awsdynamodb.PutItemInput, _ ...func(*awsdynamodb.Options)) (*awsdynamodb.PutItemOutput, error) {
						assertBananaPutItem(t, params, want)
						return &awsdynamodb.PutItemOutput{}, nil
					},
				}
			},
			wantBanana: want,
			wantErr:    nil,
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
			wantBanana: banana.Banana{},
			wantErr:    domain.ErrAlreadyExists,
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
			wantBanana: banana.Banana{},
			wantErr:    errSDK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := banana.NewRepository(tt.setupMock(t))
			got, err := repo.Create(context.Background(), want)

			assertBananaRepoResult(t, "Create", got, err, tt.wantBanana, tt.wantErr)
		})
	}
}

func TestBananaRepositoryList(t *testing.T) {
	t.Parallel()

	b1, b2, b3 := testutil.ListBananas(true)
	wantItems := []banana.Banana{b1, b2}
	page2 := []banana.Banana{b3}
	scanOutputItems := scanItems(t, wantItems)
	page2ScanItems := scanItems(t, page2)

	tests := []struct {
		name      string
		setupMock func(t *testing.T) *mockDynamoClient
		wantItems []banana.Banana
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
			wantItems: []banana.Banana{},
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
									"id": &types.AttributeValueMemberS{Value: b2.ID},
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

			repo := banana.NewRepository(tt.setupMock(t))
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
