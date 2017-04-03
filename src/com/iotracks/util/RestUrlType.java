package com.iotracks.util;

import io.netty.handler.codec.http.HttpMethod;

public enum RestUrlType {

    WS_LOG_API ("/log/", null) ,
    GET_LOGS_API ("/logs", HttpMethod.POST) ;

    private String url;
    private HttpMethod httpMethod;

    RestUrlType(String url, HttpMethod httpMethod){
        this.url = url;
        this.httpMethod = httpMethod;
    }

    public String getURL(){
        return this.url;
    }

    public HttpMethod getHttpMethod(){
        return this.httpMethod;
    }

    public static RestUrlType getByUrl(String url){
        for (RestUrlType urlType : RestUrlType.values()) {
            if(urlType.getURL().equals(url)){
                return urlType;
            }
        }
        return null;
    }

}
