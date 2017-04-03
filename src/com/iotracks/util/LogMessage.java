package com.iotracks.util;

import com.iotracks.elements.IOMessage;

import javax.json.Json;
import javax.json.JsonObject;
import javax.json.JsonReader;
import java.io.StringReader;

public class LogMessage {

    private final String PUBLISHER_ID_FIELD_NAME = "publisher_id";
    private final String LEVEL_FIELD_NAME = "level";
    private final String MESSAGE_FIELD_NAME = "message";
    private final String TIMESTAMP_FIELD_NAME = "timestamp";

    private String publisherId = "";
    private String level = "";
    private String message = "";
    private long timestamp = 0;

    public LogMessage(){}

    public LogMessage(String publisherId, String level, String message, long timestamp){
        this.publisherId = publisherId;
        this.level = level;
        this.message = message;
        this.timestamp = timestamp;
    }

    public LogMessage(IOMessage message){
        JsonReader reader = Json.createReader(new StringReader(new String(message.getContentData())));
        JsonObject logMessageJSON = reader.readObject();
        reader.close();
        if(logMessageJSON.containsKey(PUBLISHER_ID_FIELD_NAME)) {
            this.publisherId = logMessageJSON.getString(PUBLISHER_ID_FIELD_NAME);
        }
        if(logMessageJSON.containsKey(LEVEL_FIELD_NAME)) {
            this.level = logMessageJSON.getString(LEVEL_FIELD_NAME);
        }
        if(logMessageJSON.containsKey(MESSAGE_FIELD_NAME)) {
            this.message = logMessageJSON.getString(MESSAGE_FIELD_NAME);
        }
        if(logMessageJSON.containsKey(TIMESTAMP_FIELD_NAME)) {
            try{
                this.timestamp = Long.parseLong(logMessageJSON.getJsonNumber(TIMESTAMP_FIELD_NAME).toString());
            } catch(Exception e){
                System.err.println(" # Error: Invalid LONG value of '" + TIMESTAMP_FIELD_NAME + "'.");
            }
        }
    }

    /*public LogMessage(byte[] bytes) {
        int pos = 12;

        int size = ByteUtils.bytesToInteger(ByteUtils.copyOfRange(bytes, 0, 4));
        if (size > 0) {
            setPublisherId(ByteUtils.bytesToString(ByteUtils.copyOfRange(bytes, pos, pos + size)));
            pos += size;
        }

        size = ByteUtils.bytesToInteger(ByteUtils.copyOfRange(bytes, 4, 8));
        if (size > 0) {
            setLevel(ByteUtils.bytesToString(ByteUtils.copyOfRange(bytes, pos, pos + size)));
            pos += size;
        }

        size = ByteUtils.bytesToInteger(ByteUtils.copyOfRange(bytes, 8, 12));
        if (size > 0) {
            setMessage(ByteUtils.bytesToString(ByteUtils.copyOfRange(bytes, pos, pos + size)));
            pos += size;
        }

        timestamp = System.currentTimeMillis();
    }*/

    public String getPublisherId() {
        return publisherId;
    }

    /*public void setPublisherId(String publisherId) {
        this.publisherId = publisherId;
    }*/

    public String getLevel() {
        return level;
    }

    /*public void setLevel(String level) {
        this.level = level;
    }*/

    public String getMessage() {
        return message;
    }

    /*public void setMessage(String message) {
        this.message = message;
    }*/

    public long getTimestamp() {
        return timestamp;
    }

    /*public void setTimestamp(long timestamp) {
        this.timestamp = timestamp;
    }*/

    public JsonObject getJson(){
        return Json.createObjectBuilder().add(PUBLISHER_ID_FIELD_NAME, getPublisherId())
                .add(LEVEL_FIELD_NAME, getLevel())
                .add(MESSAGE_FIELD_NAME, getMessage())
                .add(TIMESTAMP_FIELD_NAME, getTimestamp())
                .build();
    }

    @Override
    public String toString() {
        return "LogMessage{ " + getJson().toString() + " }";
    }
}
