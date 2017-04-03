package com.iotracks;

import com.iotracks.api.listener.IOFabricAPIListener;
import com.iotracks.db.LogStorage;
import com.iotracks.elements.IOMessage;
import com.iotracks.util.LogMessage;

import javax.json.JsonObject;
import java.util.List;
import java.util.logging.Level;
import java.util.logging.Logger;

public class IOFabricAPIListenerImpl implements IOFabricAPIListener {

    private final Logger log = Logger.getLogger(IOFabricAPIListenerImpl.class.getName());
    private final MainLog mainLogInstance;

    public IOFabricAPIListenerImpl(MainLog mainLogInstance) {
        this.mainLogInstance = mainLogInstance;
    }

    @Override
    public void onMessages(List<IOMessage> list) {
        System.out.println("IOFabricAPIListenerImpl.onMessages");
        list.forEach(message -> {
            // infotype or infoformat should detect that it's log message and main content should be in content data as JSON to string ???
            if(message.getInfoType().equals("LOG_MESSAGE")) {
                LogStorage.addLog(new LogMessage(message));
            }
        });
    }

    @Override
    public void onMessagesQuery(long l, long l1, List<IOMessage> list) {
        System.out.println("IOFabricAPIListenerImpl.onMessagesQuery");
        /* do nothing */
    }

    @Override
    public void onError(Throwable throwable) {
        System.out.println("IOFabricAPIListenerImpl.onError");
        log.log(Level.SEVERE, "Error.", throwable);
    }

    @Override
    public void onBadRequest(String s) {
        System.out.println("IOFabricAPIListenerImpl.onBadRequest");
        log.log(Level.SEVERE, "Bad Request.", s);
    }

    @Override
    public void onMessageReceipt(String s, long l) {
        System.out.println("IOFabricAPIListenerImpl.onMessageReceipt");
        /* do nothing */
    }

    @Override
    public void onNewConfig(JsonObject jsonObject) {
        System.out.println("IOFabricAPIListenerImpl.onNewConfig");
        mainLogInstance.setConfig(jsonObject);
    }

    @Override
    public void onNewConfigSignal() {
        System.out.println("IOFabricAPIListenerImpl.onNewConfigSignal");
        mainLogInstance.updateConfig();
    }
}
