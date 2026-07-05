package dynamodb_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	awsdynamodb "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	"github.com/phides-code/go-multi-api/internal/domain"
	ddb "github.com/phides-code/go-multi-api/internal/dynamodb"
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

func scanItems(t *testing.T, bananas []domain.Banana) []map[string]types.AttributeValue {
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

	validId := uuid.NewString()
	validBanana := domain.Banana{ID: validId, Content: "ripe", CreatedOn: 12345}
	errSDK := errors.New("dynamo unavailable")

	item, err := attributevalue.MarshalMap(validBanana)
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name       string
		setupMock  func(t *testing.T) *mockDynamoClient
		wantBanana domain.Banana
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
			wantBanana: domain.Banana{},
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
			wantBanana: domain.Banana{},
			wantErr:    errSDK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := ddb.NewBananaRepository(tt.setupMock(t))
			got, err := repo.GetByID(context.Background(), validId)

			assertBananaRepoResult(t, "GetByID", got, err, tt.wantBanana, tt.wantErr)
		})
	}
}

func TestBananaRepositoryDelete(t *testing.T) {
	t.Parallel()

	validId := uuid.NewString()
	validBanana := domain.Banana{ID: validId, Content: "ripe", CreatedOn: 12345}
	errSDK := errors.New("dynamo unavailable")

	item, err := attributevalue.MarshalMap(validBanana)
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name       string
		setupMock  func(t *testing.T) *mockDynamoClient
		wantBanana domain.Banana
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
			wantBanana: domain.Banana{},
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
			wantBanana: domain.Banana{},
			wantErr:    errSDK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := ddb.NewBananaRepository(tt.setupMock(t))
			got, err := repo.Delete(context.Background(), validId)

			assertBananaRepoResult(t, "Delete", got, err, tt.wantBanana, tt.wantErr)
		})
	}
}

func TestBananaRepositoryUpdate(t *testing.T) {
	t.Parallel()

	updatedBanana := domain.Banana{ID: uuid.NewString(), Content: "updated", CreatedOn: 12345}
	errSDK := errors.New("dynamo unavailable")

	item, err := attributevalue.MarshalMap(updatedBanana)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name       string
		setupMock  func(t *testing.T) *mockDynamoClient
		wantBanana domain.Banana
		wantErr    error
	}{
		{
			name: "success",
			setupMock: func(t *testing.T) *mockDynamoClient {
				return &mockDynamoClient{
					updateItemFn: func(_ context.Context, params *awsdynamodb.UpdateItemInput, _ ...func(*awsdynamodb.Options)) (*awsdynamodb.UpdateItemOutput, error) {
						assertUpdateSets(t, params, map[string]string{
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
			wantBanana: domain.Banana{},
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
			wantBanana: domain.Banana{},
			wantErr:    errSDK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := ddb.NewBananaRepository(tt.setupMock(t))
			got, err := repo.Update(context.Background(), updatedBanana)

			assertBananaRepoResult(t, "Update", got, err, tt.wantBanana, tt.wantErr)
		})
	}
}

func TestBananaRepositoryCreate(t *testing.T) {
	t.Parallel()

	want := domain.Banana{ID: uuid.NewString(), Content: "new", CreatedOn: 12345}
	errSDK := errors.New("dynamo unavailable")

	tests := []struct {
		name       string
		setupMock  func(t *testing.T) *mockDynamoClient
		wantBanana domain.Banana
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
			wantBanana: domain.Banana{},
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
			wantBanana: domain.Banana{},
			wantErr:    errSDK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := ddb.NewBananaRepository(tt.setupMock(t))
			got, err := repo.Create(context.Background(), want)

			assertBananaRepoResult(t, "Create", got, err, tt.wantBanana, tt.wantErr)
		})
	}
}

func TestBananaRepositoryList(t *testing.T) {
	t.Parallel()

	b1 := domain.Banana{ID: uuid.NewString(), Content: "first", CreatedOn: 1}
	b2 := domain.Banana{ID: uuid.NewString(), Content: "second", CreatedOn: 2}

	cursorID := uuid.NewString()
	cursorRaw, err := json.Marshal(map[string]string{"id": cursorID})
	if err != nil {
		t.Fatal(err)
	}
	inputCursor := base64.StdEncoding.EncodeToString(cursorRaw)

	page2 := []domain.Banana{{ID: uuid.NewString(), Content: "page2", CreatedOn: 3}}
	page2ScanItems := scanItems(t, page2)

	wantItems := []domain.Banana{b1, b2}
	scanOutputItems := scanItems(t, wantItems)

	tests := []struct {
		name             string
		setupMock        func(t *testing.T) *mockDynamoClient
		wantItems        []domain.Banana
		wantNextCursorID string
		listOpts         domain.ListOptions
		wantErr          bool
		wantErrIs        error
	}{
		{
			name: "returns items",
			setupMock: func(t *testing.T) *mockDynamoClient {
				return &mockDynamoClient{
					scanFn: func(_ context.Context, params *awsdynamodb.ScanInput, _ ...func(*awsdynamodb.Options)) (*awsdynamodb.ScanOutput, error) {
						if params.Limit == nil || *params.Limit != domain.DefaultListLimit {
							t.Errorf("Limit = %v, want %v", params.Limit, domain.DefaultListLimit)
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
			wantItems: []domain.Banana{},
		},
		{
			name: "returns next cursor",
			setupMock: func(_ *testing.T) *mockDynamoClient {
				return &mockDynamoClient{
					scanFn: func(_ context.Context, _ *awsdynamodb.ScanInput, _ ...func(*awsdynamodb.Options)) (*awsdynamodb.ScanOutput, error) {
						return &awsdynamodb.ScanOutput{
							Items: scanOutputItems,
							LastEvaluatedKey: map[string]types.AttributeValue{
								"id": &types.AttributeValueMemberS{Value: b2.ID},
							},
						}, nil
					},
				}
			},
			wantItems:        wantItems,
			wantNextCursorID: b2.ID,
		},
		{
			name: "returns items for valid cursor",
			setupMock: func(_ *testing.T) *mockDynamoClient {
				return &mockDynamoClient{
					scanFn: func(_ context.Context, _ *awsdynamodb.ScanInput, _ ...func(*awsdynamodb.Options)) (*awsdynamodb.ScanOutput, error) {
						return &awsdynamodb.ScanOutput{Items: page2ScanItems}, nil
					},
				}
			},
			listOpts:  domain.ListOptions{Cursor: inputCursor},
			wantItems: page2,
		},
		{
			name: "invalid cursor",
			setupMock: func(_ *testing.T) *mockDynamoClient {
				return &mockDynamoClient{
					scanFn: func(_ context.Context, _ *awsdynamodb.ScanInput, _ ...func(*awsdynamodb.Options)) (*awsdynamodb.ScanOutput, error) {
						return nil, errors.New("scan should not be called for an invalid cursor")
					},
				}
			},
			listOpts:  domain.ListOptions{Cursor: "!!!not-base64!!!"},
			wantErr:   true,
			wantErrIs: domain.ErrInvalidCursor,
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

			repo := ddb.NewBananaRepository(tt.setupMock(t))
			page, err := repo.List(context.Background(), tt.listOpts)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.wantErrIs != nil && !errors.Is(err, tt.wantErrIs) {
					t.Fatalf("err = %v, want %v", err, tt.wantErrIs)
				}
				return
			}

			if err != nil {
				t.Fatalf("List: %v", err)
			}

			if len(page.Items) != len(tt.wantItems) {
				t.Fatalf("len(Items) = %d, want %d", len(page.Items), len(tt.wantItems))
			}

			for i := range tt.wantItems {
				if page.Items[i] != tt.wantItems[i] {
					t.Fatalf("Items[%d] = %+v, want %+v", i, page.Items[i], tt.wantItems[i])
				}
			}

			if tt.wantNextCursorID != "" {
				if page.NextCursor == "" {
					t.Fatal("expected NextCursor, got empty")
				}

				raw, err := base64.StdEncoding.DecodeString(page.NextCursor)
				if err != nil {
					t.Fatalf("decode cursor: %v", err)
				}

				var parsed map[string]string
				if err := json.Unmarshal(raw, &parsed); err != nil {
					t.Fatalf("unmarshal cursor: %v", err)
				}

				if parsed["id"] != tt.wantNextCursorID {
					t.Fatalf("cursor id = %q, want %q", parsed["id"], tt.wantNextCursorID)
				}
			} else if page.NextCursor != "" {
				t.Fatalf("unexpected NextCursor: %q", page.NextCursor)
			}
		})
	}
}
