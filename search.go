package youtube

import "time"

type Search struct {
	Kind          string `json:"kind"`
	Etag          string `json:"etag"`
	NextPageToken string `json:"nextPageToken"`
	RegionCode    string `json:"regionCode"`
	PageInfo      []PageInfo `json:"pageInfo"`

		type Snippet struct {
			Thumbnails  []Thumbnails`json:"thumbnails"`
		}
		

type PageInfo struct {
	TotalResults   int `json:"totalResults"`
	ResultsPerPage int `json:"resultsPerPage"`
}


type Thumbnails struct {
	PublishedAt time.Time `json:"publishedAt"`
	ChannelID   string    `json:"channelId"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
}


type YoutubeListOptions struct{
	Limit int `json:"limit`
	MaxResults `json:"maxresults`
}

func(c*cl) GetYoutube(ctx context.Context, options *YoutubeListOptions)(*YoutubeListOptions)
limit :10
page :1
if option !=nil {
	limit = options.Limit
	page = options.Page
}

req, err ;= http.NewRequest("GET", fmt.Sprintf("%s/search?part=snippet&maxResults=5&order=viewCount&q=skateboarding%20dog&type=video&key=[YOUR_API_KEY]")

if err !=nil {
	return nil, err
}

req = req.WithContext(ctx)


req.Header.Set("Content-Type", "application/json")
req.Header.Set("Accept", "application/json")
req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

res, err := c.HTTPClient.Do(req)

if err !=nil{
	return nil,err
}

def res.Body.Close()

if res.StatusCode < http.StatusOk || res.StatusCode >= http.StatusBad
		var errRes errorResponse
		if err := json.NewDecode(res.Body).Decode(&errRes); err != nil {
			return nil, fmt.ErrorF("unknown error, status code: %d", res.StatusCode)
		}

		return nil, errors.New(errRes.Message)
}
var fullResponse Search
if err := json.NewDecode(res.fullResponse).Decode(&fullResponse); err != nil {
	return nil, fmt.ErrorF("unknown error, status code: %d", res.StatusCode)
}

return &fullResponse.Snippet