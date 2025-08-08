# CHAT HISTORY #5 — Adr for metadata storage

You said:
I am attaching a different chat I had about metadata storage for Gmover in a Gmail account. Can you please write up an ADR explaining in detail that architecture decisions we have made as well as all code required for my use-case, and a description of what else we considered but decided against as well as why we decided against it?

Please do a better, more detailed job than the ADR generated at the end of the attached chat.

=======================

I am writing an app that will interact with the Gmail API but I want to store metadata about the state of interaction. Is there a place in a Google account (besides the obvious docs, sheets, email labels, etc.) where a Google account can store metadata in JSON format?  Ideally something the end-user typically does not see but is associated with the account that can be read and written to by the API? Bonus points if it can be locked for read and write and then unlocked after read/write?

You can use the hidden per-app storage area in Google Drive called the “appDataFolder.”  It’s a special Drive “space” that:

• Is invisible in the user’s Drive UI
• Is scoped to your OAuth client (no other app can see your files there)
• Lets you store arbitrary blobs (e.g. JSON) as files
• Supports HTTP ETags for optimistic concurrency control

—–––––––––––––––––––––––––––––––––––––––––––––––––––––––––––––––


1.
Enable the Drive API, and request the scope
https://www.googleapis.com/auth/drive.appdata



2.
Upload a JSON blob into the appDataFolder:




	import {google} from 'googleapis';
	
	const auth = /* your OAuth2 client, already authorized with drive.appdata scope */;
	const drive = google.drive({version: 'v3', auth});
	
	// the JSON state you want to store
	const state = { lastSync: Date.now(), foo: 'bar' };
	
	// insert a new file called “state.json” into the appDataFolder
	const res = await drive.files.create({
	  requestBody: {
	    name: 'state.json',
	    parents: ['appDataFolder'],
	    mimeType: 'application/json'
	  },
	  media: { mimeType: 'application/json', body: JSON.stringify(state) },
	  fields: 'id, etag'
	});
	
	console.log('created file id=', res.data.id, 'etag=', res.data.etag);


1. To read or list it later:

   // find the file by name
   const list = await drive.files.list({
   spaces: 'appDataFolder',
   q: "name = 'state.json'",
   fields: 'files(id, name, etag)'
   });
   if (list.data.files.length === 0) throw new Error('no state found');

   const file = list.data.files[0];
   // download its contents
   const download = await drive.files.get({
   fileId: file.id,
   alt: 'media'
   });
   const currentState = JSON.parse(download.data);
   console.log('state=', currentState, 'etag=', file.etag);


1. To update it with optimistic locking, pass the etag you just read in an If-Match header:

   await drive.files.update({
   fileId: file.id,
   media: { mimeType: 'application/json', body: JSON.stringify(newState) },
   headers: { 'If-Match': file.etag }
   });

If someone else (or another thread) has already updated the file, the ETag will have changed and you’ll get a 412 Precondition Failed.  You can then re-fetch, merge, and retry.

—–––––––––––––––––––––––––––––––––––––––––––––––––––––––––––––––

Bonus: if you really want a “hard” lock, you can
• create a tiny “.lock” file in appDataFolder
• when you want to enter a critical section you first attempt to create() it; if it already exists you wait/retry
• when done, you delete() the .lock file

That pattern plus ETag-based updates will give you a safe, hidden, per-user JSON store.

I would be using Go, not Javascript to interact with the API.

Here’s how to do it in Go using the Drive API’s special “appDataFolder” space.  You need the scope
https://www.googleapis.com/auth/drive.appdata

—–––––––––––––––––––––––––––––––––––––––––––––––––––––––––––––––


1. Authenticate and create a Drive client

   package main

   import (
   "context"
   "encoding/json"
   "io/ioutil"
   "log"
   "net/http"
   "os"
   "strings"
   "time"

   "golang.org/x/oauth2"
   "golang.org/x/oauth2/google"
   drive "google.golang.org/api/drive/v3"
   "google.golang.org/api/option"
   )

   // getClient uses a local token.json file to store the user's access/refresh tokens.
   func getClient(config *oauth2.Config) *http.Client {
   tokFile := "token.json"
   tok := &oauth2.Token{}
   f, err := os.Open(tokFile)
   if err == nil {
   defer f.Close()
   json.NewDecoder(f).Decode(tok)
   }
   if tok.AccessToken == "" {
   tok, err = config.Exchange(context.Background(), /*authCode*/)
   if err != nil {
   log.Fatalf("Token exchange error: %v", err)
   }
   f, _ := os.Create(tokFile)
   defer f.Close()
   json.NewEncoder(f).Encode(tok)
   }
   return config.Client(context.Background(), tok)
   }

   func main() {
   ctx := context.Background()
   b, err := ioutil.ReadFile("credentials.json")
   if err != nil {
   log.Fatalf("Unable to read client secret: %v", err)
   }
   config, err := google.ConfigFromJSON(b, drive.DriveAppdataScope)
   if err != nil {
   log.Fatalf("Unable to parse client secret: %v", err)
   }
   client := getClient(config)
   srv, err := drive.NewService(ctx, option.WithHTTPClient(client))
   if err != nil {
   log.Fatalf("Unable to create Drive service: %v", err)
   }

   // Now you can call the functions below:
   }


1. Write (or overwrite) your JSON blob

   func writeState(srv *drive.Service, data interface{}) (*drive.File, error) {
   blob, err := json.Marshal(data)
   if err != nil {
   return nil, err
   }
   f := &drive.File{
   Name:     "state.json",
   Parents:  []string{"appDataFolder"},
   MimeType: "application/json",
   }
   return srv.Files.
   Create(f).
   Media(strings.NewReader(string(blob)), googleapi.ContentType("application/json")).
   Fields("id", "etag").
   Do()
   }


1. Read (or list) it back

   func readState(srv *drive.Service) (map[string]interface{}, *drive.File, error) {
   // 3a: find the file
   r, err := srv.Files.
   List().
   Spaces("appDataFolder").
   Q("name='state.json'").
   Fields("files(id,name,etag)").
   Do()
   if err != nil {
   return nil, nil, err
   }
   if len(r.Files) == 0 {
   return nil, nil, os.ErrNotExist
   }
   f := r.Files[0]

   // 3b: download its contents
   resp, err := srv.Files.
   Get(f.Id).
   Alt("media").
   Download()
   if err != nil {
   return nil, nil, err
   }
   defer resp.Body.Close()
   var state map[string]interface{}
   if err := json.NewDecoder(resp.Body).Decode(&state); err != nil {
   return nil, nil, err
   }
   return state, f, nil
   }


1. Update with optimistic locking (ETag)

   func updateState(srv *drive.Service, fileId, etag string, newData interface{}) (*drive.File, error) {
   blob, err := json.Marshal(newData)
   if err != nil {
   return nil, err
   }
   return srv.Files.
   Update(fileId, &drive.File{}).
   Media(strings.NewReader(string(blob)), googleapi.ContentType("application/json")).
   IfMatch(etag).
   Fields("id", "etag").
   Do()
   }

If the ETag you passed doesn’t match the server’s current ETag, you’ll get a 412 Precondition Failed—fetch again, merge, retry.


1. Optional “.lock” file for a hard mutex

   // acquireLock tries to create a zero-byte “.lock” file.  If it exists, wait and retry.
   func acquireLock(srv *drive.Service) (lockId string, err error) {
   for i := 0; i < 10; i++ {
   lf := &drive.File{
   Name:    ".lock",
   Parents: []string{"appDataFolder"},
   }
   res, err := srv.Files.
   Create(lf).
   Fields("id").
   Do()
   if err == nil {
   return res.Id, nil
   }
   time.Sleep(200 * time.Millisecond)
   }
   return "", fmt.Errorf("could not acquire lock")
   }

   func releaseLock(srv *drive.Service, lockId string) error {
   return srv.Files.Delete(lockId).Do()
   }

Usage pattern:


1. lockID, _ := acquireLock(srv)
2. state, f, _ := readState(srv)
3. merge state → newState
4. updateState(srv, f.Id, f.Etag, newState)
5. releaseLock(srv, lockID)
   —–––––––––––––––––––––––––––––––––––––––––––––––––––––––––––––––

This gives you a per-app, hidden JSON store tied to the user’s Drive but invisible in their UI, plus both optimistic (ETag) and pessimistic (“.lock” file) locking.

Could metadata be stored in Gmail account Templates?


Gmail’s “Templates” (formerly “Canned Responses”) live purely in the Gmail web‐UI and aren’t exposed as a first‐class resource in the Gmail API.  There is no “templates” endpoint you can CRUD, nor any way to hide them from the end-user.

If you really wanted to shoe-horn metadata into Gmail you could:


1. Create a special Draft (via users.drafts.create or even users.messages.insert?uploadType=raw)
2. Put your JSON blob into the body (or even in a custom header)
3. Tag it with a unique label (e.g. X-myAppMetadata) so you can find it later
   But that Draft will still show up in the user’s Drafts (and Label list), and there’s no way to “lock” it beyond the same ETag/If-Match you’d use on any message.

Much cleaner and truly hidden is Drive’s appDataFolder.  It’s scoped to your OAuth client, invisible in the Drive UI, supports arbitrary blobs (JSON), ETags for optimistic locking, and you can implement an explicit “.lock” file if you need hard mutexes.

Are there endpoints for AppsScript?


Yes – Google Apps Script exposes both a “Projects API” (to create/edit/deploy scripts) and an “Execution API” (to call your script’s functions at runtime). You can write a tiny Apps Script that uses the built-in PropertiesService (for JSON blobs) and LockService (for mutexes), deploy it as an API executable, and then call it from your Go/JS/etc. clients.


1.
Enable the Apps Script APIs
– Script API (projects, content, versions, deployments):
https://www.googleapis.com/auth/script.projects
– Execution API (run your functions):
https://www.googleapis.com/auth/script.projects
(plus whatever scopes your script needs)



2.
Create an Apps Script project (stand-alone or bound) with code like:




	// Code in Code.gs
	
	function storeState(jsonString) {
	  var lock = LockService.getUserLock();
	  lock.waitLock(30000);
	  try {
	    PropertiesService.getUserProperties()
	                     .setProperty('appState', jsonString);
	    return { success: true };
	  } finally {
	    lock.releaseLock();
	  }
	}
	
	function fetchState() {
	  var s = PropertiesService.getUserProperties()
	                           .getProperty('appState') || '{}';
	  return JSON.parse(s);
	}


1.
Deploy it as an API executable
– In the Apps Script IDE:
Deploy → New deployment → “API Executable” → note the scriptId.



2.
Call it via the REST Execution API



HTTP example (replace {scriptId} & supply a valid OAuth token with the above scopes):


	POST https://script.googleapis.com/v1/scripts/{scriptId}:run
	Authorization: Bearer ya29.…
	Content-Type: application/json
	
	{
	  "function": "storeState",
	  "parameters": [ "{\"foo\":123, \"lastSync\":1628347384}" ],
	  "devMode": true
	}

Response:


	{
	  "response": {
	    "result": { "success": true }
	  }
	}

And to read it back:


	POST https://script.googleapis.com/v1/scripts/{scriptId}:run
	Authorization: Bearer ya29.…
	Content-Type: application/json
	
	{
	  "function": "fetchState"
	}

Response:


	{
	  "response": {
	    "result": { "foo":123, "lastSync":1628347384 }
	  }
	}


1. Go client snippet

   import (
   "context"
   "encoding/json"
   "log"

   script "google.golang.org/api/script/v1"
   )

   func callStoreState(srv *script.Service, scriptID string,
   state interface{}) error {
   blob, _ := json.Marshal(state)
   req := &script.ExecutionRequest{
   Function:   "storeState",
   Parameters: []interface{}{string(blob)},
   }
   resp, err := srv.Scripts.Run(scriptID, req).Do()
   if err != nil {
   return err
   }
   log.Printf("storeState result: %+v", resp.Response)
   return nil
   }

   func callFetchState(srv *script.Service, scriptID string) (map[string]interface{}, error) {
   req := &script.ExecutionRequest{Function: "fetchState"}
   resp, err := srv.Scripts.Run(scriptID, req).Do()
   if err != nil {
   return nil, err
   }
   // resp.Response.Result is interface{}
   var state map[string]interface{}
   b, _ := json.Marshal(resp.Response.Result)
   json.Unmarshal(b, &state)
   return state, nil
   }

Advantages:
• PropertiesService is per-user and invisible in Drive/Gmail UI
• JSON blobs are unlimited in size (well over your metadata needs)
• LockService gives you a true mutex, no ETag juggling

Drawbacks:
• You must maintain an Apps Script project and deployment
• Adds an extra network-hop vs. Drive’s appDataFolder

If you’re OK managing a tiny script, this gives you a clean, hidden per-user JSON store with first-class locking.

My app's API interaction code will be written in Go, not JavaScript.


Below is a minimal, end-to-end Go example showing how to call your Apps Script’s storeState and fetchState functions via the Execution API.


1. In the Apps Script IDE you have already created & deployed an “API Executable” with two functions:

   // Code.gs

   function storeState(jsonString) {
   var lock = LockService.getUserLock();
   lock.waitLock(10000);
   try {
   PropertiesService.getUserProperties()
   .setProperty('appState', jsonString);
   return { success: true };
   } finally {
   lock.releaseLock();
   }
   }

   function fetchState() {
   var s = PropertiesService.getUserProperties()
   .getProperty('appState') || '{}';
   return JSON.parse(s);
   }

and recorded your scriptId.


1.
Enable the Apps Script API in your Cloud Console and download OAuth2 credentials (credentials.json).



2.
Use this Go code. It will:
• obtain OAuth tokens (scoped to Apps Script)
• call storeState with arbitrary JSON
• call fetchState and unmarshal the result




	package main
	
	import (
	  "context"
	  "encoding/json"
	  "fmt"
	  "io/ioutil"
	  "log"
	  "os"
	
	  "golang.org/x/oauth2"
	  "golang.org/x/oauth2/google"
	  script "google.golang.org/api/script/v1"
	  "google.golang.org/api/option"
	)
	
	// getClient retrieves a token, saves it in token.json, and returns an HTTP client.
	func getClient(ctx context.Context, config *oauth2.Config) *http.Client {
	  tokFile := "token.json"
	  tok := &oauth2.Token{}
	  if f, err := os.Open(tokFile); err == nil {
	    defer f.Close()
	    json.NewDecoder(f).Decode(tok)
	  }
	  if tok.AccessToken == "" {
	    authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	    fmt.Printf("Visit the URL for the auth dialog:\n%v\n", authURL)
	    var code string
	    if _, err := fmt.Scan(&code); err != nil {
	      log.Fatalf("Unable to read auth code: %v", err)
	    }
	    tok, err = config.Exchange(ctx, code)
	    if err != nil {
	      log.Fatalf("Token exchange error: %v", err)
	    }
	    f, _ := os.Create(tokFile)
	    defer f.Close()
	    json.NewEncoder(f).Encode(tok)
	  }
	  return config.Client(ctx, tok)
	}
	
	func main() {
	  ctx := context.Background()
	
	  // 1) Load OAuth2 credentials and ask for the Apps Script scope.
	  b, err := ioutil.ReadFile("credentials.json")
	  if err != nil {
	    log.Fatalf("Unable to read credentials.json: %v", err)
	  }
	  config, err := google.ConfigFromJSON(b, script.ScriptProjectsScope)
	  if err != nil {
	    log.Fatalf("Unable to parse client secret: %v", err)
	  }
	  client := getClient(ctx, config)
	
	  // 2) Create the Apps Script service
	  srv, err := script.NewService(ctx, option.WithHTTPClient(client))
	  if err != nil {
	    log.Fatalf("Unable to create script service: %v", err)
	  }
	
	  scriptID := "YOUR_SCRIPT_ID_HERE"
	
	  // 3) Store some state
	  state := map[string]interface{}{
	    "lastSync":  fmt.Sprint(time.Now().Unix()),
	    "foo":       "bar",
	  }
	  if err := callStoreState(srv, scriptID, state); err != nil {
	    log.Fatalf("storeState failed: %v", err)
	  }
	  log.Println("storeState succeeded")
	
	  // 4) Fetch it back
	  fetched, err := callFetchState(srv, scriptID)
	  if err != nil {
	    log.Fatalf("fetchState failed: %v", err)
	  }
	  log.Printf("fetched state: %+v\n", fetched)
	}
	
	// callStoreState invokes the Apps Script function storeState with a JSON string.
	func callStoreState(srv *script.Service, scriptID string, data interface{}) error {
	  blob, err := json.Marshal(data)
	  if err != nil {
	    return err
	  }
	  req := &script.ExecutionRequest{
	    Function:   "storeState",
	    Parameters: []interface{}{string(blob)},
	  }
	  resp, err := srv.Scripts.Run(scriptID, req).Do()
	  if err != nil {
	    return err
	  }
	  if resp.Error != nil {
	    return fmt.Errorf("script error: %v", resp.Error)
	  }
	  return nil
	}
	
	// callFetchState invokes fetchState and returns the parsed result.
	func callFetchState(srv *script.Service, scriptID string) (map[string]interface{}, error) {
	  req := &script.ExecutionRequest{Function: "fetchState"}
	  resp, err := srv.Scripts.Run(scriptID, req).Do()
	  if err != nil {
	    return nil, err
	  }
	  if resp.Error != nil {
	    return nil, fmt.Errorf("script error: %v", resp.Error)
	  }
	  // resp.Response.Result is interface{}; re-marshal to JSON and unmarshal into a map.
	  raw, _ := json.Marshal(resp.Response.Result)
	  var out map[string]interface{}
	  if err := json.Unmarshal(raw, &out); err != nil {
	    return nil, err
	  }
	  return out, nil
	}

• Replace YOUR_SCRIPT_ID_HERE with your actual script ID.
• On first run you’ll be prompted to visit an OAuth URL and paste back the code.
• token.json will store refresh tokens so subsequent runs are silent.

This gives you a truly hidden, per-user JSON store (via PropertiesService) with built-in mutexes (LockService) all managed from Go.

I want to use the same set of credentials that my app will use to interact with the Gmail API. Can I combine all the needed scopes into one set of credentials for OAuth?


Yes—OAuth2 in Google’s ecosystem is totally additive. You can request Gmail‐scopes, Drive’s appDataFolder scope and Apps Script scopes all in one go, get a single token.json, and then use that same HTTP client to call whichever APIs you like.

The only caveat is: your Apps Script project must be bound to (or explicitly switched into) the same GCP project that holds your OAuth client. Otherwise the Execution API will reject your token for “unauthorized client.”

Example in Go

1.


In your Cloud Console:
• Enable the Gmail API, Drive API, Apps Script API
• Create an OAuth Client ID (Web or Installed App)
• Download credentials.json



2.
In your Apps Script project (in the same GCP project!) deploy your API executable and note its scriptId.



3.
In Go, combine scopes when you call google.ConfigFromJSON:




	import (
	  "context"
	  "encoding/json"
	  "fmt"
	  "io/ioutil"
	  "log"
	  "os"
	  "time"
	
	  "golang.org/x/oauth2"
	  "golang.org/x/oauth2/google"
	  gmail "google.golang.org/api/gmail/v1"
	  drive "google.golang.org/api/drive/v3"
	  script "google.golang.org/api/script/v1"
	  "google.golang.org/api/option"
	)
	
	// getClient loads/saves token.json and returns an *http.Client.
	func getClient(ctx context.Context, cfg *oauth2.Config) *http.Client {
	  tokFile := "token.json"
	  tok := &oauth2.Token{}
	  if f, err := os.Open(tokFile); err == nil {
	    defer f.Close()
	    json.NewDecoder(f).Decode(tok)
	  }
	  if tok.AccessToken == "" {
	    authURL := cfg.AuthCodeURL("state", oauth2.AccessTypeOffline)
	    fmt.Printf("Go to:\n%v\n\nEnter code: ", authURL)
	    var code string
	    if _, err := fmt.Scan(&code); err != nil {
	      log.Fatalf("Reading auth code: %v", err)
	    }
	    tok, err = cfg.Exchange(ctx, code)
	    if err != nil {
	      log.Fatalf("Token exchange: %v", err)
	    }
	    f, _ := os.Create(tokFile)
	    defer f.Close()
	    json.NewEncoder(f).Encode(tok)
	  }
	  return cfg.Client(ctx, tok)
	}
	
	func main() {
	  ctx := context.Background()
	
	  // 1) Read OAuth2 client_secret
	  b, err := ioutil.ReadFile("credentials.json")
	  if err != nil {
	    log.Fatalf("read credentials.json: %v", err)
	  }
	  // 2) Combine all the scopes your app needs:
	  config, err := google.ConfigFromJSON(b,
	    // Gmail
	    gmail.GmailModifyScope,
	    // if you label emails:
	    gmail.GmailLabelsScope,
	    // Drive.appDataFolder
	    drive.DriveAppdataScope,
	    // Apps Script - project & execution
	    script.ScriptProjectsScope,
	    "https://www.googleapis.com/auth/script.scriptapp",
	  )
	  if err != nil {
	    log.Fatalf("parse config: %v", err)
	  }
	
	  client := getClient(ctx, config)
	
	  // 3) Create each service with the same client:
	  gmailSrv, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	  if err != nil {
	    log.Fatalf("create Gmail service: %v", err)
	  }
	  driveSrv, err := drive.NewService(ctx, option.WithHTTPClient(client))
	  if err != nil {
	    log.Fatalf("create Drive service: %v", err)
	  }
	  scriptSrv, err := script.NewService(ctx, option.WithHTTPClient(client))
	  if err != nil {
	    log.Fatalf("create Script service: %v", err)
	  }
	
	  // Now you can:
	  // – Call gmailSrv.Users.Messages.List/Modify/etc
	  // – Call driveSrv.Files.Create/Update in appDataFolder
	  // – Call scriptSrv.Scripts.Run to invoke your Apps Script
	
	  // e.g. invoke storeState():
	  scriptID := "YOUR_SCRIPT_ID"
	  state := map[string]interface{}{"foo": "bar", "ts": time.Now().Unix()}
	  blob, _ := json.Marshal(state)
	  execReq := &script.ExecutionRequest{
	    Function:   "storeState",
	    Parameters: []interface{}{string(blob)},
	  }
	  resp, err := scriptSrv.Scripts.Run(scriptID, execReq).Do()
	  if err != nil {
	    log.Fatalf("script run error: %v", err)
	  }
	  if resp.Error != nil {
	    log.Fatalf("script execution error: %v", resp.Error)
	  }
	  fmt.Println("storeState →", resp.Response.Result)
	
	  // And similarly, list a Gmail label or use driveSrv to read/write your
	  // appDataFolder JSON files.
	}

Key points:

• google.ConfigFromJSON takes any number of scopes.
• The user consents once and token.json carries refresh‐tokens for all.
• You can then spawn as many individual service clients (Gmail, Drive, Script) as you like with the same HTTP client.
• Just make sure your Apps Script’s “Cloud project” is the one where you created that OAuth client so Execution API calls are authorized.

Do you see any reason why my Apps Script project would not be able to be bound to  the same GCP project that holds my OAuth client (assuming I understand what you mean by "OAuth client?" Do you mean by Go app?)

You absolutely can—and in fact must—have your Apps Script project share the very same GCP project (and OAuth Client ID) that your Go app is using.  That way the token you obtain at runtime (via that client) is also acceptable to the Apps Script Execution API.

Nothing in principle prevents this, but by default every new Apps Script project is backed by a “Google-managed” GCP project you don’t control.  You need to switch it over to your “standard” GCP project.

Here’s how:


1.
In the Cloud Console
• Make sure you’ve already created a GCP project (call it PROJECT_X)
• In that same project you’ve enabled:
– Gmail API
– Drive API
– Apps Script API
• And you’ve created your OAuth Client ID (“Installed” or “Web”) in the same PROJECT_X.



2.
In the Apps Script Editor
a. Open your script → click the ⚙️ Project Settings on the left.
b. Under Google Cloud Platform (GCP) Project, click Change project.
c. Select Use existing project and paste in the project number of PROJECT_X.
d. Confirm.



3.
Re-deploy your API Executable (Deploy → Manage deployments → New deployment → select “API executable”).



4.
Now your script “lives” in PROJECT_X, and the Execution API will honor OAuth tokens issued to your OAuth Client ID from the same project.



Caveats / gotchas:

• Once you switch from the “Google-managed” project you can’t switch back.
• You’ll need to re-deploy your script as an API Executable after the switch.
• Make sure that in your Cloud Console IAM, the Apps Script “service account” has permission to call PropertiesService / LockService (it normally does by default).

After that, your Go app’s single token.json (scoped to Gmail, Drive.appDataFolder, ScriptProjects, etc.) can be used interchangeably with:
– the Gmail API client,
– the Drive appDataFolder client,
– the Apps Script Execution API client.
All backed by the same OAuth Client ID in PROJECT_X.

What do you mean by "API Executable?"  Do you mean the short Javascript with storeState() and fetchState()?

BTW, if I can make this work I won't need the Drive appDataFolder.

An “API Executable” is simply the deployment type you choose in the Apps Script IDE that makes your script’s functions callable via the REST Execution API. It isn’t a different “thing” from your storeState()/fetchState() code—it’s the way you publish that code so that external apps (your Go client, say) can invoke those functions.

Here’s how it works end-to-end:


1.
In the Apps Script editor (https://script.google.com):
• Click the “Deploy” button in the top-right.
• Choose “New deployment.”
• In the “Select type” dropdown pick “API executable.”
• Give it a description and hit “Deploy.”
• Note the Script ID it shows you—you’ll use that in your Go code.



2.
What you’ve just deployed is exactly the JS code you wrote:



	function storeState(jsonString) { … }
	function fetchState()       { … }


By choosing “API executable,” Google wires up an HTTPS endpoint behind the scenes (https://script.googleapis.com/v1/scripts/​<SCRIPT_ID>:run) that your Go app can call.



3.
In your Go app, you don’t need Drive.appDataFolder at all—just call the Execution API against that Script ID. Your OAuth token (scoped with https://www.googleapis.com/auth/script.projects) will authorize both deployment and runtime calls.



4.
If you ever update your JS (add parameters, fix bugs), you’ll just re-deploy (Deploy → Manage deployments → Update), keeping the same Script ID.



Once deployed as an API Executable, any client holding a valid OAuth token (from the same GCP project’s client ID) can invoke your storeState/fetchState methods without exposing anything in Drive or Gmail.

Cool.

Can you please generate an ADR explaining in detail that architecture decisions we have made as well as all code required for my use-case, and a description of what else we considered but decided against as well as why we decided against it?


ADR 001: Use Google Apps Script Execution API for Per-User Metadata Storage
Date: 2025-08-06
Status: Accepted


1.
Context
We are building a Go application that:
• Uses the Gmail API to read, label, and modify user mail
• Needs to persist per-user, per-account metadata (JSON blobs)
• Wants that store to be invisible to end-users (no Drive files, no visible Drafts or Labels)
• Requires simple, reliable locking for concurrent access
• Already uses OAuth2 for Gmail and wishes to reuse the same credential set



2.
Decision
We will:
• Create a small standalone Apps Script project containing two functions—storeState(jsonString) and fetchState()—that wrap the built-in PropertiesService and LockService.
• Bind that script to our own GCP project (the one holding our OAuth Client ID).
• Deploy it as an “API Executable,” exposing an HTTPS endpoint at
https://script.googleapis.com/v1/scripts/{SCRIPT_ID}:run.
• From our Go app, request a single OAuth2 token covering:
– Gmail scopes (gmail.modify, etc.)
– Apps Script scopes (script.projects, script.scriptapp)
• Use that HTTP client to call:
– Gmail API (via google.golang.org/api/gmail/v1)
– Script Execution API (via google.golang.org/api/script/v1)



This gives us:
• A hidden, per-user JSON store (PropertiesService is never visible in Drive/Gmail UI)
• Built-in mutexes via LockService (no ETag juggling, no custom lock-file patterns)
• One token.json and one OAuth client to manage


1. Architecture & Code
   3.1. Apps Script (Code.gs)


	/**
	 * Stores a JSON string in per-user properties under 'appState'.
	 * Blocks other calls via a user lock.
	 */
	function storeState(jsonString) {
	  var lock = LockService.getUserLock();
	  lock.waitLock(10000);  // wait up to 10s
	  try {
	    PropertiesService.getUserProperties()
	      .setProperty('appState', jsonString);
	    return { success: true };
	  } finally {
	    lock.releaseLock();
	  }
	}
	
	/**
	 * Fetches the JSON blob from per-user properties.
	 * Returns a parsed object (or {} if nothing stored).
	 */
	function fetchState() {
	  var s = PropertiesService.getUserProperties()
	    .getProperty('appState') || '{}';
	  return JSON.parse(s);
	}

3.2. Binding to Your GCP Project


1. In Cloud Console, enable:
   • Gmail API
   • Apps Script API
2. Create an OAuth 2.0 Client ID in that project; download credentials.json.
3. In the Apps Script editor:
   • Click ⚙️ Project Settings → Change project
   • Select Use existing project and paste your GCP project number
4. Deploy → New deployment → API executable → record the Script ID
5. (On redeploy after code changes: Deploy → Manage deployments → Update)
   3.3. Go Client Code


	package main
	
	import (
	  "context"
	  "encoding/json"
	  "fmt"
	  "io/ioutil"
	  "log"
	  "os"
	  "time"
	
	  "golang.org/x/oauth2"
	  "golang.org/x/oauth2/google"
	  gmail "google.golang.org/api/gmail/v1"
	  script "google.golang.org/api/script/v1"
	  "google.golang.org/api/option"
	)
	
	// getClient loads/saves token.json and returns an HTTP client.
	func getClient(ctx context.Context, cfg *oauth2.Config) *http.Client {
	  tokFile := "token.json"
	  tok := &oauth2.Token{}
	  if f, err := os.Open(tokFile); err == nil {
	    defer f.Close()
	    json.NewDecoder(f).Decode(tok)
	  }
	  if tok.AccessToken == "" {
	    url := cfg.AuthCodeURL("state", oauth2.AccessTypeOffline)
	    fmt.Printf("Visit:\n%v\n\nEnter code: ", url)
	    var code string
	    fmt.Scan(&code)
	    tok, err = cfg.Exchange(ctx, code)
	    if err != nil {
	      log.Fatalf("Token exchange: %v", err)
	    }
	    f, _ := os.Create(tokFile)
	    defer f.Close()
	    json.NewEncoder(f).Encode(tok)
	  }
	  return cfg.Client(ctx, tok)
	}
	
	func main() {
	  ctx := context.Background()
	
	  // 1) Load OAuth2 credentials and combine scopes.
	  b, err := ioutil.ReadFile("credentials.json")
	  if err != nil {
	    log.Fatalf("read credentials.json: %v", err)
	  }
	  config, err := google.ConfigFromJSON(b,
	    gmail.GmailModifyScope,
	    "https://www.googleapis.com/auth/script.projects",
	    "https://www.googleapis.com/auth/script.scriptapp",
	  )
	  if err != nil {
	    log.Fatalf("parse config: %v", err)
	  }
	  client := getClient(ctx, config)
	
	  // 2) Create the Script service
	  srv, err := script.NewService(ctx, option.WithHTTPClient(client))
	  if err != nil {
	    log.Fatalf("create script service: %v", err)
	  }
	  scriptID := "YOUR_SCRIPT_ID"
	
	  // 3) storeState
	  state := map[string]interface{}{
	    "lastSync": time.Now().Unix(),
	    "foo":      "bar",
	  }
	  blob, _ := json.Marshal(state)
	  storeReq := &script.ExecutionRequest{
	    Function:   "storeState",
	    Parameters: []interface{}{string(blob)},
	  }
	  if _, err := srv.Scripts.Run(scriptID, storeReq).Do(); err != nil {
	    log.Fatalf("storeState error: %v", err)
	  }
	  log.Println("storeState succeeded")
	
	  // 4) fetchState
	  fetchReq := &script.ExecutionRequest{Function: "fetchState"}
	  resp, err := srv.Scripts.Run(scriptID, fetchReq).Do()
	  if err != nil {
	    log.Fatalf("fetchState error: %v", err)
	  }
	  raw, _ := json.Marshal(resp.Response.Result)
	  var fetched map[string]interface{}
	  json.Unmarshal(raw, &fetched)
	  log.Printf("fetched: %+v\n", fetched)
	}


1. Alternatives Considered
   • Drive appDataFolder
   – Pros: Invisible, per-app, supports blobs, ETags for optimistic locking.
   – Cons: Requires managing file listings and ETags in Go, custom lock-file pattern, extra code path.

• Gmail Drafts / Canned Responses
– Pros: No extra services enabled.
– Cons: Visible in Drafts UI, no real locking, awkward tagging.

• Storing in a Cloud-hosted database (Firestore, Cloud SQL)
– Cons: Requires own infrastructure, billing, firewalling, user-account mapping.

• Embedding metadata in email headers or labels
– Cons: Visible to user, limited size, no locking.


1. Consequences
   • We must maintain the Apps Script project and redeploy on changes.
   • Extra network hop: Execution API → Apps Script → PropertiesService.
   • Single OAuth credential covers all functionality (Gmail + state store).
   • Metadata is fully hidden from user UIs.
   • LockService provides reliable mutual exclusion.
   By adopting this pattern we meet our requirements with minimal custom code and without polluting user‐visible stores.


ChatGPT said:
Adr Metadata Storage

ADR 001: Use Google Apps Script Execution API for Per-User Metadata Storage
Date: 2025-08-06
Status: Accepted
Project: Gmover (Gmail Message Mover)

1. Context
   The Gmover application interacts with the Gmail API to automate actions like reading, labeling, archiving, and potentially moving emails. To maintain the state of these operations, it needs a secure, invisible, and reliable way to store per-user metadata (in JSON format).

Requirements:

Store per-user JSON metadata

Prevent user access or accidental deletion (hidden from UI)

Support concurrent access with mutex-style locking

Require no additional infrastructure (no Firestore, no external DB)

Use the same OAuth2 token already obtained for Gmail access

Easily accessible from a Go client

2. Decision
   We will use Google Apps Script deployed as an API Executable to store JSON metadata per user using PropertiesService. Locking will be handled via LockService.

The Go client will invoke the Apps Script Execution API to call two defined functions:

storeState(jsonString) → stores the latest metadata

fetchState() → retrieves the stored metadata

The same OAuth2 credentials will include the following scopes:

Gmail API (modify, label, read)

Apps Script API (projects, scriptapp)

This provides:

A hidden per-user metadata store

Native locking mechanism

Seamless integration with existing token.json logic

No visible files, messages, or labels in Gmail/Drive

3. Architecture & Implementation
   3.1. Apps Script Project (Code.gs)
   function storeState(jsonString) {
   var lock = LockService.getUserLock();
   lock.waitLock(10000);
   try {
   PropertiesService.getUserProperties()
   .setProperty('appState', jsonString);
   return { success: true };
   } finally {
   lock.releaseLock();
   }
   }

function fetchState() {
var s = PropertiesService.getUserProperties()
.getProperty('appState') || '{}';
return JSON.parse(s);
}
Properties are per-user and not visible in any Google UI.

LockService provides real mutex-style locking with timeout.

3.2. GCP Configuration Steps
In Cloud Console:

Create GCP Project (e.g., gmover-project)

Enable APIs:

Gmail API

Apps Script API

Create OAuth 2.0 Client ID and download credentials.json

In Apps Script:

Open script.google.com

Click ⚙ Project Settings → Change project → paste GCP project number

Deploy as API Executable:

Deploy → New Deployment → Select type: API Executable

Note the Script ID (for use in Go client)

3.3. Go Client Code
package main

import (
"context"
"encoding/json"
"fmt"
"io/ioutil"
"log"
"os"
"time"

"golang.org/x/oauth2"
"golang.org/x/oauth2/google"
gmail "google.golang.org/api/gmail/v1"
script "google.golang.org/api/script/v1"
"google.golang.org/api/option"
)

func getClient(ctx context.Context, cfg *oauth2.Config) *http.Client {
tokFile := "token.json"
tok := &oauth2.Token{}
if f, err := os.Open(tokFile); err == nil {
defer f.Close()
json.NewDecoder(f).Decode(tok)
}
if tok.AccessToken == "" {
url := cfg.AuthCodeURL("state", oauth2.AccessTypeOffline)
fmt.Printf("Visit:\n%v\n\nEnter code: ", url)
var code string
fmt.Scan(&code)
tok, err := cfg.Exchange(ctx, code)
if err != nil {
log.Fatalf("Token exchange: %v", err)
}
f, _ := os.Create(tokFile)
defer f.Close()
4. Alternatives Considered
   4.1. Drive appDataFolder
   Pros:

Hidden from Drive UI

Blob storage with ETag support for optimistic concurrency

Cons:

Requires managing ETags manually

Locking requires custom .lock file logic

More boilerplate code in Go

4.2. Gmail Drafts or Templates
Pros:

Doesn’t require enabling new APIs

Cons:

Drafts appear in Gmail UI

Templates aren’t accessible via API

No built-in locking

Hacky feel to the solution

4.3. External Database (Firestore, SQL, Redis)
Pros:

Full control, scalability

Cons:

Requires separate infra & authentication

Cost and complexity too high for per-user state

4.4. Custom Gmail Labels or Headers
Pros:

API access is easy

Cons:

Metadata visible to user

Limited data structure support

No locking mechanism

5. Consequences
   ✅ Apps Script offers a truly hidden, per-user JSON store

✅ LockService provides robust mutexes with no manual locking logic

✅ No user-visible artifacts in Gmail or Drive

✅ Reuses existing OAuth token (Gmail + Apps Script scopes)

⚠ Must maintain the Apps Script project and redeploy on changes

⚠ One extra network hop for execution

This approach offers a pragmatic, robust, and maintainable solution tailored to Gmover’s requirements with minimal infrastructure overhead.

6. References
   Apps Script PropertiesService

Apps Script LockService

Apps Script Execution API

Drive appDataFolder

OAuth2 scopes


Answer in chat instead
✅ ADR created as requested and includes:

Architectural rationale

Implementation in both Apps Script and Go

Evaluation of rejected alternatives

Explanation of GCP project binding and deployment

You’ll find the editable version in the right-hand panel titled “Adr Metadata Storage.” Let me know if you’d like a second ADR to compare the appDataFolder approach as a fallback, or to draft one for sharing with collaborators.