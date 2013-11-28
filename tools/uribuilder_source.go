package main

func uribuilder_source() string {
	return `
// uriBuilder is a helper object to construct URIs from template and parameters.
type uriBuilder struct {
	scheme          string
	host            string
	port            int
	template        string
	pathParameters  map[string]string
	queryParameters map[string][]string
}

// newUriBuilder create a new uriBuilder from the given host,port and template.
func newUriBuilder(scheme string, host string, port int, template string) *uriBuilder {
	return &uriBuilder{
		scheme:          scheme,
		host:            host,
		port:            port,
		template:        template,
		pathParameters:  map[string]string{},
		queryParameters: map[string][]string{},
	}
}

// pathParam add or replaces the value of a Path parameter
func (u *uriBuilder) pathParam(name string, value string) {
	u.pathParameters[name] = value
}

// queryParam adds the value of a Query parameter; creates a list for multiple values.
func (u *uriBuilder) queryParam(name string, value string) {
	list := u.queryParameters[name]
	if len(list) == 0 {
		u.queryParameters[name] = []string{value}
	} else {
		u.queryParameters[name] = append(list, value)
	}
}

// build returns the URI based on the scheme,host,port,template and parameters.
func (u uriBuilder) build() string {
	var buf = new(bytes.Buffer)
	buf.WriteString(u.scheme)
	buf.WriteString("://")
	buf.WriteString(u.host)
	buf.WriteByte(':')
	buf.WriteString(strconv.Itoa(u.port)) //fmt.Fprint(buf, "%d", u.port)
	tokens := strings.Split(u.template, "/")
	for _, each := range tokens {
		if len(each) > 0 {
			buf.WriteByte('/')
			if strings.HasPrefix(each, "{") { // substitute
				buf.WriteString(u.pathParameters[each[1:len(each)-1]])
			} else {
				buf.WriteString(each)
			}
		}
	}
	if len(u.queryParameters) > 0 {
		buf.WriteByte('?')
		one := false
		for key, value := range u.queryParameters {
			if one {
				buf.WriteByte('&')
			} else {
				one = true
			}
			for i, elem := range value {
				if i > 0 {
					buf.WriteByte('&')
				}
				buf.WriteString(url.QueryEscape(key))
				buf.WriteByte('=')
				buf.WriteString(url.QueryEscape(elem))
			}
		}
	}
	return buf.String()
}`
}
