package framework

type Requeueable struct {
	requeue bool
}

func (in *Requeueable) SetNeedsRequeue(requeue bool) {
	in.requeue = requeue
}

func (in *Requeueable) NeedsRequeue() bool {
	return in.requeue
}
