package main

type serverFormLabel struct {
	name   string
	secret bool
}

func serverFormLabels() []serverFormLabel {
	return []serverFormLabel{
		{name: "配置名称"},
		{name: "服务器码"},
		{name: "服务器密码", secret: true},
	}
}
