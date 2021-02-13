package users_test

import (
	"testing"

	"github.com/curious-kitten/scratch-post/pkg/administration/users"
)

func TestUser_Validate(t *testing.T) {
	tests := []struct {
		name    string
		u       *users.User
		wantErr bool
	}{
		{
			name: "Valid user",
			u: &users.User{
				Name:     "Test user",
				Username: "testuser94",
				Email:    "test.user@email.com",
				Password: "4Passw0rd.F0r.T3st!!",
			},
			wantErr: false,
		},
		{
			name: "Missing name",
			u: &users.User{
				Username: "testuser94",
				Email:    "test.user@email.com",
				Password: "4Passw0rd.F0r.T3st!!",
			},
			wantErr: true,
		},
		{
			name: "Invalid email",
			u: &users.User{
				Name:     "Test user",
				Username: "testuser94",
				Email:    "test user@email.com",
				Password: "4Passw0rd.F0r.T3st!!",
			},
			wantErr: true,
		},
		{
			name: "Invalid password",
			u: &users.User{
				Name:     "Test user",
				Username: "testuser94",
				Email:    "testuser@email.com",
				Password: "asda",
			},
			wantErr: true,
		},
		{
			name: "Invalid Username",
			u: &users.User{
				Name:     "Test user",
				Username: "testuser94@",
				Email:    "testuser@email.com",
				Password: "asda",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.u.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("User.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEmail_IsValid(t *testing.T) {
	tests := []struct {
		name string
		e    users.Email
		want bool
	}{
		{
			name: "Space in address",
			e:    "test user@email.com",
			want: false,
		},
		{
			name: "no @",
			e:    "abc-mail.com",
			want: false,
		},
		{
			name: "Contains :",
			e:    "test:@email.com",
			want: false,
		},
		{
			name: "Starts with @",
			e:    "@test@email.com",
			want: false,
		},
		{
			name: "Ends with @",
			e:    "test@email.com@",
			want: false,
		},
		{
			name: "Ends with @",
			e:    "test@email.com@",
			want: false,
		},
		{
			name: "no prefix",
			e:    "@email.com",
			want: false,
		},
		{
			name: "no domain",
			e:    "test@",
			want: false,
		},
		{
			name: "valid",
			e:    "test@email.com",
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.e.IsValid(); got != tt.want {
				t.Errorf("Email.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPassword_IsValid(t *testing.T) {
	tests := []struct {
		name string
		p    users.Password
		want bool
	}{
		{
			name: "Valid password",
			p:    "4Passw0rd.F0r.T3st!!",
			want: true,
		},
		{
			name: "Password with no capital letters",
			p:    "4passw0rd.f0r.t3st!!",
			want: false,
		},
		{
			name: "Password with no numbers",
			p:    "Password.For.Test!!",
			want: false,
		},
		{
			name: "Password with no punctuation",
			p:    "4Passw0rdF0rT3st",
			want: false,
		},
		{
			name: "Password to short",
			p:    "4.Pass",
			want: false,
		},
		{
			name: "Password with space",
			p:    "4Passw0rd .F0r.T3st!!",
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.IsValid(); got != tt.want {
				t.Errorf("Password.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUsername_IsValid(t *testing.T) {
	tests := []struct {
		name string
		u    users.Username
		want bool
	}{
		{
			name: "Valid",
			u:    "test_username",
			want: true,
		},
		{
			name: "username contains @",
			u:    "testuser94@",
			want: false,
		},
		{
			name: "username contains @",
			u:    "testuser94@",
			want: false,
		},
		{
			name: "username is less then 3",
			u:    "ab",
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.u.IsValid(); got != tt.want {
				t.Errorf("Username.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}
