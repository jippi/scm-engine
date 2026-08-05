package backstage

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"strings"

	"github.com/datolabs-io/go-backstage/v3"
	go_backstage "github.com/datolabs-io/go-backstage/v3"
	"github.com/jippi/scm-engine/pkg/scm"
	slogctx "github.com/veqryn/slog-context"
	"golang.org/x/oauth2"
)

type Client struct {
	wrapped *go_backstage.Client
}

// NewClient creates a new Backstage client with an optional bearer token
func NewClient(ctx context.Context, baseURL string, token string, httpClient *http.Client) (*Client, error) {
	if token != "" && httpClient == nil {
		httpClient = oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{
			AccessToken: token,
			TokenType:   "Bearer",
		}))
	}

	if baseURL == "" {
		return nil, errors.New("baseURL is required")
	}

	backstageClient, err := go_backstage.NewClient(baseURL, "default", httpClient)
	if err != nil {
		return nil, err
	}

	return &Client{wrapped: backstageClient}, nil
}

func (c *Client) Client() *go_backstage.Client {
	return c.wrapped
}

// GetEntityOwner returns the entity reference for the entity owner.
//
// Entity References https://backstage.io/docs/features/software-catalog/references#string-references
// Filtering https://backstage.io/docs/features/software-catalog/software-catalog-api#filtering
func (c *Client) GetEntityOwner(ctx context.Context, filters ...string) (*EntityReference, error) {
	systems, response, err := c.wrapped.Catalog.Entities.List(ctx, &backstage.ListEntityOptions{
		Filters: filters,
		Fields: []string{
			"spec.owner",
		},
	})
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get entity owner: %s", response.Status)
	}

	if len(systems) == 0 {
		return nil, errors.New("No system found in Backstage catalog")
	}

	if len(systems) > 1 {
		slogctx.Warn(ctx, "Multiple systems found for filters, defaulting to first", slog.Any("filters", filters))
	}

	// spec.owner can be a group or a user
	// https://backstage.io/docs/features/software-catalog/descriptor-format#specowner-required
	entityRef := systems[0].Spec["owner"]

	entityRefStr, ok := entityRef.(string) // add type assertion check
	if !ok {
		return nil, fmt.Errorf("Expected spec.owner to be a string: %v", entityRef)
	}

	parsedEntityRef, err := parseEntityReference(entityRefStr)
	if err != nil {
		return nil, err
	}

	return parsedEntityRef, nil
}

// GetUser returns a Backstage user entity.
func (c *Client) GetUser(ctx context.Context, userEntityRef *EntityReference) (*go_backstage.UserEntityV1alpha1, error) {
	user, response, err := c.wrapped.Catalog.Users.Get(ctx, userEntityRef.Name, userEntityRef.Namespace)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Failed to get user: %s", response.Status)
	}

	return user, nil
}

// ListGroupMembers returns all Backstage entities that are members of a group.
func (c *Client) ListGroupMembers(ctx context.Context, groupEntityRef *EntityReference) ([]go_backstage.Entity, error) {
	var users []go_backstage.Entity

	ref := groupEntityRef.ToString()

	users, response, err := c.wrapped.Catalog.Entities.List(ctx, &backstage.ListEntityOptions{
		Filters: []string{
			// https://backstage.io/docs/features/software-catalog/well-known-relations/#memberof-and-hasmember
			"kind=user,relations.memberof=" + ref,
		},
	})
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Failed to get group members: %s", response.Status)
	}

	return users, nil
}

func (c *Client) GetOwnersForGitLabProject(ctx context.Context, projectName string) ([]scm.Actor, error) {
	// (kind=system AND metadata.name=?) OR (...)
	entityRef, err := c.GetEntityOwner(ctx,
		"kind=system,metadata.name="+projectName,
		"kind=system,metadata.annotations.gitlab.com/project="+projectName,
	)
	if err != nil {
		return nil, err
	}

	var actors []scm.Actor

	if entityRef.IsUser() {
		userEntity, err := c.GetUser(ctx, entityRef)
		if err != nil {
			return nil, err
		}

		actors = convertBackstageEntitiesToGitLabActors(userEntity.Entity)
	} else if entityRef.IsGroup() {
		entities, err := c.ListGroupMembers(ctx, entityRef)
		if err != nil {
			return nil, err
		}

		actors = convertBackstageEntitiesToGitLabActors(entities...)
	}

	return actors, nil
}

// https://backstage.io/docs/features/software-catalog/references
type EntityReference struct {
	Name      string
	Kind      string
	Namespace string
}

func (e *EntityReference) IsGroup() bool {
	return e.Kind == "group"
}

func (e *EntityReference) IsUser() bool {
	return e.Kind == "user"
}

func (e *EntityReference) ToString() string {
	var builder strings.Builder

	if e.Kind != "" {
		builder.WriteString(e.Kind)
		builder.WriteString(":")
	}

	if e.Namespace != "" {
		builder.WriteString(e.Namespace)
		builder.WriteString("/")
	}

	builder.WriteString(e.Name)

	return builder.String()
}

// parseEntityReference parses a Backstage entity reference string into an EntityReference struct.
//
// Kind and Namespace are optional, e.g. [<kind>:][<namespace>/]<name>
func parseEntityReference(entityRef string) (*EntityReference, error) {
	ref := &EntityReference{}

	// use regex to parse the entity reference
	re := regexp.MustCompile(`^(?:(?P<kind>[^:]+):)?(?P<namespace>[^/]+/)?(?P<name>.+)$`)
	match := re.FindStringSubmatch(entityRef)

	if len(match) == 0 {
		return nil, fmt.Errorf("invalid entity reference format: %s", entityRef)
	}

	ref.Kind = match[1]
	ref.Namespace = match[2]
	ref.Name = match[3]

	// the namespace in the regex match includes a trailing slash, so we need to remove it
	if strings.HasSuffix(ref.Namespace, "/") {
		ref.Namespace = ref.Namespace[:len(ref.Namespace)-1]
	}

	// default to the default namespace
	if ref.Namespace == "" {
		ref.Namespace = "default"
	}

	// default to group if kind is not set
	if ref.Kind == "" {
		ref.Kind = "group"
	}

	return ref, nil
}

// Helper function to convert Backstage user entities to GitLab actors
func convertBackstageEntitiesToGitLabActors(entities ...go_backstage.Entity) []scm.Actor {
	actors := make([]scm.Actor, 0, len(entities))

	for _, entity := range entities {
		userID, ok := entity.Metadata.Annotations["gitlab.com/user_id"]
		if !ok {
			continue
		}

		actors = append(actors, scm.Actor{
			ID: userID,
		})
	}

	return actors
}
