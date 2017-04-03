package com.iotracks;

import com.iotracks.api.IOFabricClient;
import com.iotracks.db.LogStorage;
import com.iotracks.rest.RestApiServer;
import com.iotracks.util.LogMessage;
import io.netty.util.internal.StringUtil;

import javax.json.JsonObject;
import java.util.concurrent.Executors;
import java.util.concurrent.ScheduledExecutorService;
import java.util.concurrent.ScheduledFuture;
import java.util.concurrent.TimeUnit;
import java.util.logging.Logger;

/**
 * @author iotracks
 * Main class
 */
public class MainLog {

    private final static Logger log = Logger.getLogger(MainLog.class.getName());

    private final String LOG_MESSAGE_TIME_LIMIT_PROP_NAME = "cleanup_interval";

    private static Object fetchConfigLock = new Object();
    private static JsonObject config = null;
    private static String containerId = "TEST_LOG_CONTAINER";

    private static IOFabricClient ioFabricClient;
    private static IOFabricAPIListenerImpl listener;

    private static ScheduledExecutorService cleanScheduler;
    private static ScheduledFuture<?> scheduledFuture;
    private static Long logMessageTimeLimit = 60 * 60 * 24 * 2L; // 2 days

    public static void main(String[] args) throws Exception {

        MainLog instance = new MainLog();
        LogStorage.create();

        Thread restApiServerThread = new Thread(new RestApiServer());
        restApiServerThread.start();

        cleanScheduler = Executors.newScheduledThreadPool(4);

        if (args.length > 0 && args[0].startsWith("--id=")) {
            containerId = args[0].substring(args[0].indexOf('=') + 1);
        } else {
            containerId = System.getenv("SELFNAME");
        }

        if (StringUtil.isNullOrEmpty(containerId)) {
            System.err.println("Container Id is not specified. Please, use --id=XXXX parameter or set the id as SELFNAME=XXXX environment property");
        } else {
            String ioFabricHost = System.getProperty("iofabric_host", "iofabric");
            int ioFabricPort = 54321;
            try {
                ioFabricPort = Integer.parseInt(System.getProperty("iofabric_port", "54321"));
            } catch (Exception e) {
            /* default value 54321 will be used */
            }

            ioFabricClient = new IOFabricClient(ioFabricHost, ioFabricPort, containerId);
            listener = new IOFabricAPIListenerImpl(instance);

            updateConfig();

            try {
                ioFabricClient.openControlWebSocket(listener);
            } catch (Exception e) {
                System.err.println("Unable to open Control WebSocket to ioFog.\n" + e.getMessage());
            }

            try {
                ioFabricClient.openMessageWebSocket(listener);
            } catch (Exception e) {
                System.err.println("Unable to open Message WebSocket to ioFog.\n" + e.getMessage());
            }
        }
    }

    public void setConfig(JsonObject configObject) {
        System.out.println("MainLog.setConfig: " + configObject);
        config = configObject;
        synchronized (fetchConfigLock) {
            fetchConfigLock.notifyAll();
        }
        if(config.containsKey(LOG_MESSAGE_TIME_LIMIT_PROP_NAME)) {
            try {
                logMessageTimeLimit = Long.parseLong(config.getJsonNumber(LOG_MESSAGE_TIME_LIMIT_PROP_NAME).toString());
            } catch (Exception e) {
                System.err.println("Bad_Request: Can't parse value for " + LOG_MESSAGE_TIME_LIMIT_PROP_NAME + "property: " + e.getMessage());
            }
        }
        startCleaner();
    }

    public static void updateConfig(){
        System.out.println("MainLog.updateConfig");
        config = null;
        stopCleaner();
        try {
            while (config == null) {
                ioFabricClient.fetchContainerConfig(listener);
                synchronized (fetchConfigLock) {
                    fetchConfigLock.wait(1000);
                }
            }
        } catch (Exception e) {
            System.err.println("Error fetching config. " + e.getMessage());
        }
    }

    private static void stopCleaner(){
        System.out.println("MainLog.stopCleaner");
        if(scheduledFuture != null) {
            scheduledFuture.cancel(false);
        }
    }

    private static void startCleaner(){
        System.out.println("MainLog.startCleaner");
        scheduledFuture = cleanScheduler.scheduleWithFixedDelay(new MessageCleaner(), 0, logMessageTimeLimit, TimeUnit.MILLISECONDS);
    }

    private static class MessageCleaner implements Runnable{

        public MessageCleaner(){}

        @Override
        public void run() {
            System.out.println("MessageCleaner. running Cleaner");
            Long currentTime = System.currentTimeMillis();
            LogStorage.deleteMessages(null, currentTime - logMessageTimeLimit);
        }
    }
}
