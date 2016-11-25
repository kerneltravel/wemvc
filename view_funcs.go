package wemvc

import (
	"html/template"
	"net/http"
)

func include_view(path string, ctx map[string]interface{}) interface{} {
	var ns *NsSection
	if ctx != nil {
		if nsInterface, ok := ctx["Namespace"]; ok && nsInterface != nil {
			nsTmp,ok := nsInterface.(*NsSection)
			if ok && nsTmp != nil && len(nsTmp.name) > 0 {
				ns = nsTmp
			}
		}
	}
	if ns == nil {
		bytes, err :=  RenderView(path, ctx)
		if err == nil {
			return  template.HTML(bytes)
		}
		panic(err)
	} else {
		bytes,err := ns.RenderView(path, ctx)
		if err == nil {
			return template.HTML(bytes)
		}
		panic(err)
	}
}

func req_query(req *http.Request, key string) string {
	if req == nil || len(key) == 0 {
		return ""
	}
	return req.URL.Query().Get(key)
}

func req_form(req *http.Request, key string) string {
	if req == nil || len(key) == 0 {
		return ""
	}
	return req.Form.Get(key)
}

func req_header(req *http.Request, key string) string {
	if req == nil || len(key) == 0 {
		return ""
	}
	return req.Header.Get(key)
}

func req_postForm(req *http.Request, key string) string {
	if req == nil || len(key) == 0 {
		return ""
	}
	return req.PostForm.Get(key)
}

func req_host(req *http.Request) string {
	if req == nil {
		return ""
	}
	return req.URL.Host
}

func cache_view(cache *CacheManager, key string) interface{} {
	if cache == nil || len(key) == 0 {
		return nil
	}
	return cache.Get(key)
}

func session_view(session SessionStore, key interface{}) interface{} {
	if session == nil || key == nil {
		return nil
	}
	return session.Get(key)
}