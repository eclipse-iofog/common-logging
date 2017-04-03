package com.iotracks.rest;

import com.iotracks.util.LogMessage;
import com.iotracks.db.LogStorage;
import com.iotracks.util.RestUrlType;
import io.netty.buffer.ByteBuf;
import io.netty.handler.codec.http.*;

import javax.json.*;
import java.io.StringReader;
import java.util.*;
import java.util.concurrent.Callable;

import static io.netty.handler.codec.http.HttpResponseStatus.OK;
import static io.netty.handler.codec.http.HttpVersion.HTTP_1_1;


class HttpRequestHandler implements Callable {

    private final String TIMEFRAME_START_PARAM_NAME = "timeframestart";
    private final String TIMEFRAME_END_PARAM_NAME = "timeframeend";
    private final String PUBLISHERS_PARAM_NAME = "publishers";
    private final String BATCH_NUMBER_PARAM_NAME = "batch_number";

    private final String RESPONSE_STATUS_PROP_NAME = "status";
    private final String RESPONSE_SIZE_PROP_NAME = "size";
    private final String RESPONSE_LOGS_PROP_NAME = "logs";

    private final int maxLimit = 1000;

    private final FullHttpRequest req;
    private ByteBuf bytesData;
    private RestUrlType urlType;

    public HttpRequestHandler(FullHttpRequest req, ByteBuf bytesData, RestUrlType urlType){
        this.req = req;
        this.bytesData = bytesData;
        this.urlType = urlType;
    }

    @Override
    public Object call() throws Exception {
        System.out.println("HttpRequestHandler.call");
        HttpHeaders headers = req.headers();

        if (req.getMethod() != urlType.getHttpMethod()) {
            return sendErrorResponse(Collections.singleton(" # Error: Incorrect HTTP method type."));
        }

        if(!(headers.get(HttpHeaders.Names.CONTENT_TYPE).equals("application/json"))){
            return sendErrorResponse(Collections.singleton(" # Error: Incorrect HTTP headers."));
        }

        ByteBuf msgBytes = req.content();
        String requestBody = msgBytes.toString(io.netty.util.CharsetUtil.US_ASCII);
        JsonReader reader = Json.createReader(new StringReader(requestBody));
        JsonObject jsonObject = reader.readObject();

        switch (urlType) {
            case GET_LOGS_API:
                return handleGetLogsRequest(jsonObject);
        }
        return sendErrorResponse(Collections.singleton(" # Error: Unhandled request call."));
    }

    private void checkField(JsonObject jsonObject, String fieldName, Set<String > errors){
        if(!jsonObject.containsKey(fieldName)){
            errors.add(" # Error: Missing input field '" + fieldName +  "'.");
        }
    }

    private void parseLongField(JsonObject jsonObject, String fieldName, Set<String > errors){
        try{
            if(jsonObject.containsKey(fieldName)) {
                Long.parseLong(jsonObject.getJsonNumber(fieldName).toString());
            }
        } catch(Exception e){
            errors.add(" # Error: Invalid value of '" + fieldName + "'.");
        }
    }

    private void validateMessageQuery(JsonObject jsonObject, Set<String> errors){
        checkField(jsonObject, PUBLISHERS_PARAM_NAME, errors);
        checkField(jsonObject, TIMEFRAME_START_PARAM_NAME, errors);
        checkField(jsonObject, TIMEFRAME_END_PARAM_NAME, errors);

        parseLongField(jsonObject, TIMEFRAME_START_PARAM_NAME, errors);
        parseLongField(jsonObject, TIMEFRAME_END_PARAM_NAME, errors);
    }

    private FullHttpResponse sendErrorResponse(Set<String> errors){
        errors.forEach(error -> bytesData.writeBytes(error.getBytes()));
        return new DefaultFullHttpResponse(HTTP_1_1, HttpResponseStatus.BAD_REQUEST, bytesData);
    }

    private FullHttpResponse sendResponse(){
        System.out.println("HttpRequestHandler.sendResponse");
        FullHttpResponse res = new DefaultFullHttpResponse(HTTP_1_1, OK, bytesData);
        HttpHeaders.setContentLength(res, bytesData.readableBytes());
        return res;
    }

    private JsonObject buildMessagesResponse(List<LogMessage> messages, JsonObject jsonObject) {
        System.out.println("HttpRequestHandler.buildMessagesResponse");
        JsonArrayBuilder messagesBuilder = Json.createArrayBuilder();
        boolean pagination = false;
        int startIndex = 0, maxIndex = messages.size(), batchNumber = 1 ;
        if(messages.size() >= maxLimit) {
            pagination = true;
            if( jsonObject.containsKey(BATCH_NUMBER_PARAM_NAME)) {
                try {
                    batchNumber = jsonObject.getInt(BATCH_NUMBER_PARAM_NAME);
                } catch (Exception e) { /* default value = 1 will be used */ }
            }
            int maxBatchNumber = messages.size()/maxLimit;
            if(messages.size() % maxLimit !=0) {
                maxBatchNumber++;
            }
            if(batchNumber > maxBatchNumber) {
                batchNumber = maxBatchNumber;
            } else if (batchNumber <= 0) {
                batchNumber = 1;
            }
            startIndex = (batchNumber - 1) * maxLimit;
            maxIndex = batchNumber * maxLimit;
            if (maxIndex > messages.size()) {
                maxIndex = messages.size();
            }
        }
        for(int i = startIndex; i < maxIndex; i++) {
            messagesBuilder.add(messages.get(i).getJson());
        }
        JsonObjectBuilder jsonBuilder = Json.createObjectBuilder()
                .add(RESPONSE_STATUS_PROP_NAME, "okay")
                .add(RESPONSE_SIZE_PROP_NAME, messages.size())
                .add(RESPONSE_LOGS_PROP_NAME, messagesBuilder);
        if(pagination) {
            jsonBuilder.add(BATCH_NUMBER_PARAM_NAME, batchNumber);
        }
        return jsonBuilder.build();
    }

    private FullHttpResponse handleGetLogsRequest(JsonObject jsonObject){
        System.out.println("HttpRequestHandler.handleGetLogsRequest");
        Set<String> errors = new HashSet<>();
        validateMessageQuery(jsonObject, errors);
        if(!errors.isEmpty()) {
            return sendErrorResponse(errors);
        }
        JsonArray publishersArray = jsonObject.getJsonArray(PUBLISHERS_PARAM_NAME);
        List<String> publishers = new ArrayList<>(publishersArray.size());
        publishersArray.forEach(jsonValue -> publishers.add(jsonValue.toString()));
        List<LogMessage> messages = LogStorage.getMessages( publishers, Long.parseLong(jsonObject.getJsonNumber(TIMEFRAME_START_PARAM_NAME).toString()), Long.parseLong(jsonObject.getJsonNumber(TIMEFRAME_END_PARAM_NAME).toString()));
        if(messages != null) {
            bytesData.writeBytes(buildMessagesResponse(messages, jsonObject).toString().getBytes());
            return sendResponse();
        } else {
            return sendErrorResponse(Collections.singleton("No log messages found for provided parameters."));
        }
    }
}
