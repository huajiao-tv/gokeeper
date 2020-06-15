package discovery

//元组信息
type MD map[string]string

//md拼接
func (md MD) Join(p MD) {
	for k, v := range p {
		md[k] = v
	}
}
