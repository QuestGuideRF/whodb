package envconfig
import "testing"
func TestGetDefaultDatabaseCredentialsParsesEnv(t *testing.T) {
	t.Setenv("WHODB_POSTGRES", `[{"host":"db.local","user":"alice","password":"secret","database":"app"}]`)
	creds := GetDefaultDatabaseCredentials("postgres")
	if len(creds) != 1 {
		t.Fatalf("expected one credential parsed from env, got %d", len(creds))
	}
	if creds[0].Hostname != "db.local" || creds[0].Username != "alice" || creds[0].Database != "app" {
		t.Fatalf("unexpected credentials parsed: %+v", creds[0])
	}
}
func TestFindAllDatabaseCredentialsFallback(t *testing.T) {
	t.Setenv("WHODB_MYSQL", "")
	t.Setenv("WHODB_MYSQL_1", `{"host":"mysql.local","user":"bob","password":"pw","database":"northwind"}`)
	creds := GetDefaultDatabaseCredentials("mysql")
	if len(creds) != 1 {
		t.Fatalf("expected fallback credentials to be discovered, got %d", len(creds))
	}
	if creds[0].Hostname != "mysql.local" || creds[0].Username != "bob" {
		t.Fatalf("unexpected fallback credential: %+v", creds[0])
	}
}