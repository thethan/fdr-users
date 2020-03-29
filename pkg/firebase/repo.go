package firebase

import (
	"cloud.google.com/go/firestore"
	"context"
	"errors"
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/thethan/fdr-users/pkg/users"

	"go.elastic.co/apm"
)

type Repo struct {
	logger      log.Logger
	firestore *firestore.Client
}

func NewFirebaseRepository(logger log.Logger, firestore *firestore.Client) Repo {
	return Repo{
		logger:      logger,
		firestore: firestore,
	}
}

const AccessKey = "access_token"

func (r *Repo) GetCredentialInformation(ctx context.Context, uid string) (users.User, error) {
	span, ctx := apm.StartSpan(ctx, "GetCredentialInformation", "db.firebase.init")
	defer span.End()

	docuRef := r.firestore.Doc(fmt.Sprintf("users/%s", uid))
	if docuRef == nil {
		level.Error(r.logger).Log("message", "error in getting firestore docuref", "error", errors.New("yeah no docuref"))

		return users.User{}, errors.New("connect ")
	}

	snapShot, err := r.getDocumentReference(ctx, docuRef)
	if err != nil {
		level.Error(r.logger).Log("message", "error in getting firestore snapshot", "error", err)

		return users.User{}, err
	}

	if !snapShot.Exists() {
		level.Error(r.logger).Log("message", "snapshot did not exist", "error", err)
		return users.User{}, err
	}

	var accessKey string
	data := snapShot.Data()
	keyInterface, ok := data[AccessKey]
	if !ok {
		level.Error(r.logger).Log("message", "data did not have access key", )
		return users.User{}, errors.New("goth did not have access key")
	}

	accessKey, ok = keyInterface.(string)
	if !ok {
		_ = level.Error(r.logger).Log("message", "access key is not a string ", )
		return users.User{}, errors.New("access key was not a string")
	}

	return users.User{AccessToken: accessKey}, nil
}

func (r *Repo) getDocumentReference(ctx context.Context, docuRef *firestore.DocumentRef) (*firestore.DocumentSnapshot, error) {
	span, ctx := apm.StartSpan(ctx, "GetCredentialInformation", "db.firebase.query")
	defer span.End()

	return docuRef.Get(ctx)
}
