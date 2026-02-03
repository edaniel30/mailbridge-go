# Push Notifications

Monitor Gmail mailbox changes in real-time using Google Cloud Pub/Sub.

> **Setup required**: [OAuth2 configuration](../gmail/GMAIL.md#setup-oauth2)

## What You Get

- Real-time notifications when mailbox changes
- No polling required (efficient streaming connection)
- Track new messages, deletions, and label changes

## Architecture

```
Your App → WatchMailbox() → Gmail API
                              ↓
                          Pub/Sub Topic
                              ↓
Your App ← GetHistory() ← Notifications
```

## Quick Setup

### 1. Enable Pub/Sub

```bash
# Via gcloud CLI
gcloud services enable pubsub.googleapis.com
gcloud pubsub topics create gmail-notifications
gcloud pubsub subscriptions create gmail-subscription --topic=gmail-notifications

# Grant Gmail permission to publish
gcloud pubsub topics add-iam-policy-binding gmail-notifications \
  --member=serviceAccount:gmail-api-push@system.gserviceaccount.com \
  --role=roles/pubsub.publisher
```

**Web Console**: Pub/Sub � Create Topic (`gmail-notifications`) � Permissions � Add Principal: `gmail-api-push@system.gserviceaccount.com` with role `Pub/Sub Publisher`

### 2. Authentication for Pub/Sub Client

**Development**:
```bash
gcloud auth application-default login
```

**Production** (create service account):
1. IAM & Admin → Service Accounts → Create
2. Name: `gmail-pubsub-client`
3. Role: `Pub/Sub Subscriber`
4. Create Key (JSON) → Download
5. `export GOOGLE_APPLICATION_CREDENTIALS="/path/to/key.json"`

## API Usage

### Activate Notifications

```go
watchResp, err := client.WatchMailbox(ctx, &core.WatchRequest{
    TopicName: "projects/my-project/topics/gmail-notifications",
    LabelIDs:  []string{"INBOX"},  // Optional: filter by labels
})
// Expires in 7 days - must renew
```

### Listen for Changes

```go
// Create Pub/Sub client (separate from Gmail client)
pubsubClient, _ := pubsub.NewClient(ctx, projectID)
subscriber := pubsubClient.Subscriber("gmail-subscription")

// Streaming connection (not polling)
subscriber.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
    var notification struct {
        EmailAddress string `json:"emailAddress"`
        HistoryID    uint64 `json:"historyId"`
    }
    json.Unmarshal(msg.Data, &notification)

    // Fetch what changed
    history, _ := client.GetHistory(ctx, &core.HistoryRequest{
        StartHistoryID: fmt.Sprintf("%d", lastHistoryID),
    })

    // Process changes
    for _, record := range history.History {
        for _, added := range record.MessagesAdded {
            fmt.Printf("New message: %s\n", added.Message.ID)
        }
    }

    msg.Ack()
})
```

### Stop Notifications

```go
err := client.StopWatch(ctx)
```

## GetHistory Explained

**Why needed?** Pub/Sub notification only says "something changed" but doesn't include details.

```
Notification: { historyID: 3548 }  → Just an ID
         ↓
GetHistory(from: 3441, to: 3548)   ← Fetch details
         ↓
Response: {
  MessagesAdded: [...],      → New emails
  MessagesDeleted: [...],    → Deleted emails
  LabelsAdded: [...],        → Label changes
  LabelsRemoved: [...]
}
```

## Watch Lifecycle

- **Duration**: 7 days
- **Renewal**: Call `StopWatch()` then `WatchMailbox()` before expiration
- **Best practice**: Renew every 6 days via cron job

## Complete Example

See working code: [`examples/gmail-notifications`](../../examples/gmail-notifications/)

```bash
export GMAIL_CLIENT_ID="..."
export GMAIL_CLIENT_SECRET="..."
export GOOGLE_CLOUD_PROJECT="your-project"
export PUBSUB_TOPIC="projects/your-project/topics/gmail-notifications"

# Place your service-account-key.json in examples/gmail/ and set absolute path
export GOOGLE_APPLICATION_CREDENTIALS="$PWD/examples/gmail/service-account-key.json"

cd examples/gmail/gmail-notifications
go run main.go start  # Send yourself an email to test!
```

**Note:** The service account key should be placed in `examples/gmail/service-account-key.json`. Use absolute path for `GOOGLE_APPLICATION_CREDENTIALS` or relative path from where you run the command.

## Common Issues

**"Permission denied"** → Gmail doesn't have publisher role on topic
```bash
gcloud pubsub topics get-iam-policy gmail-notifications
```

**"could not find default credentials"** → Missing auth
```bash
gcloud auth application-default login
```

**"Resource not found"** → Subscription name mismatch (check your code vs actual subscription name)

**"No changes found"** → Normal for old/duplicate notifications

## Resources

- [Complete Example](../../examples/gmail-notifications/)
- [Gmail Push Docs](https://developers.google.com/gmail/api/guides/push)
- [Pub/Sub Docs](https://cloud.google.com/pubsub/docs)
