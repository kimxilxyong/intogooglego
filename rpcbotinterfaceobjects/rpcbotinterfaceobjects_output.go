package rpcbotinterfaceobjects

type BotOutput struct {
	Content string
}

func (this *BotOutput) GetContent() string {
	return this.Content
}

func (this *BotOutput) SetContent(NewContent string) {
	this.Content = NewContent
}
