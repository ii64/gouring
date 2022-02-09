package gouring

// Option
type Option func(p *IOUringParams)

// SQThread option
func SQThread(cpu, idleMS uint32) Option {
	return func(p *IOUringParams) {
		p.SQThreadCPU = cpu
		p.SQThreadIdle = idleMS
	}
}
