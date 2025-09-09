package chat

type PrivateMsg struct {
	from *Client
	to   string
	text string
}

type renameReq struct {
	client  *Client
	newNick string
	resp    chan error
}

type whoReq struct {
	resp chan []string
}
