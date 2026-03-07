package service

import (
	"context"
	"errors"
	"testing"

	"project-management/internal/model"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type stubAuthRepo struct {
	findByEmailFn func(ctx context.Context, email string) (model.User, error)
	createFn      func(ctx context.Context, user *model.User) error
	getByIDFn     func(ctx context.Context, id uint) (model.User, error)
}

func (s stubAuthRepo) FindByEmail(ctx context.Context, email string) (model.User, error) {
	return s.findByEmailFn(ctx, email)
}
func (s stubAuthRepo) Create(ctx context.Context, user *model.User) error {
	return s.createFn(ctx, user)
}
func (s stubAuthRepo) GetByID(ctx context.Context, id uint) (model.User, error) {
	return s.getByIDFn(ctx, id)
}

type stubPasswordManager struct {
	hashFn    func(password string) (string, error)
	compareFn func(hash, password string) error
}

func (s stubPasswordManager) Hash(password string) (string, error) { return s.hashFn(password) }
func (s stubPasswordManager) Compare(hash, password string) error  { return s.compareFn(hash, password) }

type stubTokenIssuer struct {
	issueFn func(user model.User) (string, error)
}

func (s stubTokenIssuer) Issue(user model.User) (string, error) { return s.issueFn(user) }

func TestAuthServiceRegister(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		svc := &authService{
			repo: stubAuthRepo{
				findByEmailFn: func(ctx context.Context, email string) (model.User, error) {
					if email != "user@example.com" {
						t.Fatalf("email = %q", email)
					}
					return model.User{}, gorm.ErrRecordNotFound
				},
				createFn: func(ctx context.Context, user *model.User) error {
					user.ID = 10
					if user.Name != "Alice" || user.PasswordHash != "hashed" {
						t.Fatalf("unexpected user: %+v", user)
					}
					return nil
				},
			},
			hasher: stubPasswordManager{hashFn: func(password string) (string, error) {
				if password != "secret1" {
					t.Fatalf("password = %q", password)
				}
				return "hashed", nil
			}},
			tokens: stubTokenIssuer{issueFn: func(user model.User) (string, error) {
				if user.ID != 10 {
					t.Fatalf("user id = %d", user.ID)
				}
				return "jwt", nil
			}},
		}

		user, token, err := svc.Register(ctx, RegisterInput{Email: " User@Example.com ", Password: "secret1", Name: " Alice "})
		if err != nil {
			t.Fatalf("Register error = %v", err)
		}
		if token != "jwt" || user.Email != "user@example.com" || user.Name != "Alice" {
			t.Fatalf("unexpected result: user=%+v token=%q", user, token)
		}
	})

	t.Run("email in use", func(t *testing.T) {
		svc := &authService{
			repo: stubAuthRepo{findByEmailFn: func(ctx context.Context, email string) (model.User, error) {
				return model.User{ID: 1}, nil
			}},
		}
		_, _, err := svc.Register(ctx, RegisterInput{Email: "used@example.com"})
		if !errors.Is(err, ErrEmailInUse) {
			t.Fatalf("err = %v, want ErrEmailInUse", err)
		}
	})

	t.Run("hash failure", func(t *testing.T) {
		svc := &authService{
			repo: stubAuthRepo{findByEmailFn: func(ctx context.Context, email string) (model.User, error) {
				return model.User{}, gorm.ErrRecordNotFound
			}},
			hasher: stubPasswordManager{hashFn: func(password string) (string, error) { return "", errors.New("hash failed") }},
		}
		_, _, err := svc.Register(ctx, RegisterInput{Email: "user@example.com", Password: "secret1"})
		if err == nil || err.Error() != "hash failed" {
			t.Fatalf("err = %v", err)
		}
	})

	t.Run("find existing internal error", func(t *testing.T) {
		svc := &authService{
			repo: stubAuthRepo{findByEmailFn: func(ctx context.Context, email string) (model.User, error) {
				return model.User{}, errors.New("db down")
			}},
		}
		_, _, err := svc.Register(ctx, RegisterInput{Email: "user@example.com"})
		if err == nil || err.Error() != "db down" {
			t.Fatalf("err = %v", err)
		}
	})

	t.Run("create failure", func(t *testing.T) {
		svc := &authService{
			repo: stubAuthRepo{
				findByEmailFn: func(ctx context.Context, email string) (model.User, error) {
					return model.User{}, gorm.ErrRecordNotFound
				},
				createFn: func(ctx context.Context, user *model.User) error { return errors.New("insert failed") },
			},
			hasher: stubPasswordManager{hashFn: func(password string) (string, error) { return "hashed", nil }},
		}
		_, _, err := svc.Register(ctx, RegisterInput{Email: "user@example.com", Password: "secret1"})
		if err == nil || err.Error() != "insert failed" {
			t.Fatalf("err = %v", err)
		}
	})

	t.Run("token issue failure", func(t *testing.T) {
		svc := &authService{
			repo: stubAuthRepo{
				findByEmailFn: func(ctx context.Context, email string) (model.User, error) {
					return model.User{}, gorm.ErrRecordNotFound
				},
				createFn: func(ctx context.Context, user *model.User) error { return nil },
			},
			hasher: stubPasswordManager{hashFn: func(password string) (string, error) { return "hashed", nil }},
			tokens: stubTokenIssuer{issueFn: func(user model.User) (string, error) { return "", errors.New("token failed") }},
		}
		_, _, err := svc.Register(ctx, RegisterInput{Email: "user@example.com", Password: "secret1"})
		if err == nil || err.Error() != "token failed" {
			t.Fatalf("err = %v", err)
		}
	})
}

func TestAuthServiceLogin(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		svc := &authService{
			repo: stubAuthRepo{findByEmailFn: func(ctx context.Context, email string) (model.User, error) {
				return model.User{ID: 5, Email: email, PasswordHash: "hashed"}, nil
			}},
			hasher: stubPasswordManager{compareFn: func(hash, password string) error {
				if hash != "hashed" || password != "secret1" {
					t.Fatalf("compare inputs = %q %q", hash, password)
				}
				return nil
			}},
			tokens: stubTokenIssuer{issueFn: func(user model.User) (string, error) { return "jwt", nil }},
		}
		user, token, err := svc.Login(ctx, LoginInput{Email: " USER@example.com ", Password: "secret1"})
		if err != nil {
			t.Fatalf("Login error = %v", err)
		}
		if token != "jwt" || user.Email != "user@example.com" {
			t.Fatalf("unexpected result: user=%+v token=%q", user, token)
		}
	})

	t.Run("compare failure", func(t *testing.T) {
		svc := &authService{
			repo: stubAuthRepo{findByEmailFn: func(ctx context.Context, email string) (model.User, error) {
				return model.User{PasswordHash: "hashed"}, nil
			}},
			hasher: stubPasswordManager{compareFn: func(hash, password string) error { return bcrypt.ErrMismatchedHashAndPassword }},
		}
		_, _, err := svc.Login(ctx, LoginInput{Email: "user@example.com", Password: "wrong"})
		if !errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			t.Fatalf("err = %v", err)
		}
	})

	t.Run("find by email failure", func(t *testing.T) {
		svc := &authService{
			repo: stubAuthRepo{findByEmailFn: func(ctx context.Context, email string) (model.User, error) {
				return model.User{}, errors.New("boom")
			}},
		}
		_, _, err := svc.Login(ctx, LoginInput{Email: "user@example.com", Password: "secret1"})
		if err == nil || err.Error() != "boom" {
			t.Fatalf("err = %v", err)
		}
	})

	t.Run("token issue failure", func(t *testing.T) {
		svc := &authService{
			repo: stubAuthRepo{findByEmailFn: func(ctx context.Context, email string) (model.User, error) {
				return model.User{ID: 2, Email: email, PasswordHash: "hashed"}, nil
			}},
			hasher: stubPasswordManager{compareFn: func(hash, password string) error { return nil }},
			tokens: stubTokenIssuer{issueFn: func(user model.User) (string, error) { return "", errors.New("token failed") }},
		}
		_, _, err := svc.Login(ctx, LoginInput{Email: "user@example.com", Password: "secret1"})
		if err == nil || err.Error() != "token failed" {
			t.Fatalf("err = %v", err)
		}
	})
}

func TestAuthServiceGetByID(t *testing.T) {
	svc := &authService{repo: stubAuthRepo{getByIDFn: func(ctx context.Context, id uint) (model.User, error) {
		if id != 3 {
			t.Fatalf("id = %d", id)
		}
		return model.User{ID: 3, Email: "a@example.com"}, nil
	}}}
	user, err := svc.GetByID(context.Background(), 3)
	if err != nil {
		t.Fatalf("GetByID error = %v", err)
	}
	if user.Email != "a@example.com" {
		t.Fatalf("user = %+v", user)
	}
}

func TestAuthServiceGetByIDError(t *testing.T) {
	svc := &authService{repo: stubAuthRepo{getByIDFn: func(ctx context.Context, id uint) (model.User, error) {
		return model.User{}, errors.New("boom")
	}}}
	_, err := svc.GetByID(context.Background(), 3)
	if err == nil || err.Error() != "boom" {
		t.Fatalf("err = %v", err)
	}
}

func TestPasswordManager(t *testing.T) {
	pm := passwordManager{}
	hash, err := pm.Hash("secret1")
	if err != nil {
		t.Fatalf("Hash error = %v", err)
	}
	if err := pm.Compare(hash, "secret1"); err != nil {
		t.Fatalf("Compare error = %v", err)
	}

	t.Run("hash too long", func(t *testing.T) {
		longPassword := string(make([]byte, 73))
		_, err := pm.Hash(longPassword)
		if err == nil {
			t.Fatal("expected hash error for long password")
		}
	})
}

func TestJWTTokenIssuer(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	issuer := jwtTokenIssuer{}
	token, err := issuer.Issue(model.User{ID: 1, Email: "a@example.com"})
	if err != nil || token == "" {
		t.Fatalf("token=%q err=%v", token, err)
	}
}

func TestNewAuthServiceConstructors(t *testing.T) {
	repo := stubAuthRepo{
		findByEmailFn: func(ctx context.Context, email string) (model.User, error) {
			return model.User{}, gorm.ErrRecordNotFound
		},
		createFn:  func(ctx context.Context, user *model.User) error { return nil },
		getByIDFn: func(ctx context.Context, id uint) (model.User, error) { return model.User{}, nil },
	}

	if svc := NewAuthService(repo); svc == nil {
		t.Fatal("NewAuthService returned nil")
	}
	if svc := NewAuthServiceWithDeps(repo, stubPasswordManager{hashFn: func(password string) (string, error) { return "h", nil }, compareFn: func(hash, password string) error { return nil }}, stubTokenIssuer{issueFn: func(user model.User) (string, error) { return "t", nil }}); svc == nil {
		t.Fatal("NewAuthServiceWithDeps returned nil")
	}
}
