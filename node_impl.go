// Code generated by irgen; DO NOT EDIT.

package ahm

type Proc struct {
	Name, Title string
	Children    []Node
}
type Text struct {
	Text string
}

func (Node *Proc) FeedTo(consumer NodeConsumer) { consumer.Proc(Node.Name, Node.Title, Node.Children) }
func (Node *Text) FeedTo(consumer NodeConsumer) { consumer.Text(Node.Text) }
