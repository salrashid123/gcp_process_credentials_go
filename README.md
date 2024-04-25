## Process Credentials for GCP Client Library  - golang

Google Cloud Credential provider which allows sourcing credentials from an external process.

Essentially, its a credential source which allows the delegation of acquiring GCP `access_tokens` to arbitrary binaries you have access to at runtime.

The arbitrary binary would use whatever means it has available (kerberos, ldap, saml-cli, etc) to get a GCP `access_token`.  

From there, the token is given surfaced as a refreshable credential source you can directly use with a GCP library.

This is similar to several systems that provide such delegation.

* Kubernetes kubectl [credential plugin](https://kubernetes.io/docs/reference/access-authn-authz/authentication/#client-go-credential-plugins)
* [AWS Process Credentials](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-sourcing-external.html)


Needless to say, use this after very careful consideration:  this library will attempt to execute a binary on the system where its run (ofcource the process running using the library would need access to run the script anyway)

>> **NOTE** these samples are NOT supported by google; its just something done on a weekend...

---

### Implementations

As its a weekend project, caveat emptor.  The code is alpha quality and I didn't have time to push it to maven central, npm, etc. 

If you want it there, please review the code, provide suggestions and improvements here

* `golang`: [https://github.com/salrashid123/gcp_process_credentials_go](https://github.com/salrashid123/gcp_process_credentials_go)
* `python`: [https://github.com/salrashid123/gcp_process_credentials_py](https://github.com/salrashid123/gcp_process_credentials_py)
* `java`: [https://github.com/salrashid123/gcp_process_credentials_java](https://github.com/salrashid123/gcp_process_credentials_java)
* `node`: [https://github.com/salrashid123/gcp_process_credentials_node](https://github.com/salrashid123/gcp_process_credentials_node)


See the "examples" folder in each

---

### Binary Response Contract

Each library above will invoke a binary, pass it some args and env var.

The response back from the binary must

be valid JSON in the form

```json
{
  "access_token": "ya29....",
  "expires_in": 3600,
  "token_type": "Bearer"
}
```

* `access_token`: your access token
* `expires_in`: how many seconds this token is valid for
* `token_type`:  usually just a bearer token

### Quickstart

For a quick example, the following will read a token file and use that for credentials:

```golang
import "github.com/salrashid123/gcp_process_credentials_go"

	extTokenSource, err := sal.ExternalTokenSource(
		&sal.ExternalTokenConfig{
			Command: "/usr/bin/cat",
			Env:     []string{"foo=bar"},
			Args:    []string{"/tmp/token.txt"},
		},
	)

	tok, err := extTokenSource.Token()

	sts := oauth2.ReuseTokenSource(tok, extTokenSource)
	storageClient, err := storage.NewClient(ctx, option.WithTokenSource(sts))
```

ofcourse the file  here `/tmp/token.txt` **must** be the json file format described above

### Parser Interface 

If your binary does not provide the exact json format, your can define a parser interface to 'translate' the credential for you.

For example,  `gcloud auth print-access-token` returns just the access token with an annoying newline character from stdout.

You an provide an interface to do the translation like this:

```golang
	extTokenSource, err := sal.ExternalTokenSource(
		&sal.ExternalTokenConfig{
			Command: "gcloud",
			Env:     []string{"foo=bar"},
			Args:    []string{"auth", "print-access-token"},
			Parser: func(b []byte) (sal.ExternalTokenResponse, error) {
				ret := &sal.ExternalTokenResponse{
					Token:     string(b),
					ExpiresIn: 3600,
					TokenType: "Bearer",
				}
				return *ret, nil
			},
		},
	)
```

### Injecting tokens vs Wrapped Credentials

You might be asking..._why cant' i just run the binary on my own in code, get the token an inject it as a credential like this?_?

Well, the, token is _not_ refreshable and your client library will need to manage that.  On the other hand, if you use this library, it will automatically refresh the token by calling the binary when its nearing expiration

---

Other References [AWS->GCP Process Credential Plugin](https://github.com/salrashid123/awscompat#process-credentials)




