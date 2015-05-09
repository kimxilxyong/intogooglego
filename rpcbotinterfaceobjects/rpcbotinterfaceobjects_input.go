package rpcbotinterfaceobjects

type BotInput struct {
	Content string
}

func (this *BotInput) GetContent() string {
	return this.Content
}

func (this *BotInput) SetContent(NewContent string) {
	this.Content = NewContent
}
