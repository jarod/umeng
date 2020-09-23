

```go
import (
    "context"
    "github.com/jarod/umeng/upush"
)

func main() {
    client := upush.NewClient(appKey, appMasterSecret)
    uploadResult, err := client.Upload(context.TODO(), &upush.UploadParam{
        Content:
    })
}
```