package firebase

import (
	"context"
	firebase "firebase.google.com/go"
	"github.com/go-kit/kit/log"
	"github.com/stretchr/testify/assert"
	"google.golang.org/api/option"
	"testing"
)

func TestFirebase(t *testing.T) {
	ctx := context.Background()

	logger := log.NewNopLogger()
	sa := option.WithCredentialsFile("../../serviceAccountKey.json")
	app, err := firebase.NewApp(ctx, nil, sa)
	if err != nil  {
		t.FailNow()
	}
	firestoreClient, err := app.Firestore(ctx)
	if err != nil  {
		t.FailNow()
	}

	repo := NewFirebaseRepository(logger, firestoreClient)

	assert.Nil(t, err)

	assert.NotEqual(t, "", user.AccessToken, )
}
