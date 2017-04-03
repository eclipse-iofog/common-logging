package com.iotracks.rest;

import com.iotracks.db.LogStorage;
import com.iotracks.elements.IOMessage;
import com.iotracks.util.LogMessage;
import com.iotracks.util.RestUrlType;
import com.iotracks.utils.ByteUtils;
import io.netty.buffer.ByteBuf;
import io.netty.buffer.Unpooled;
import io.netty.channel.ChannelFuture;
import io.netty.channel.ChannelFutureListener;
import io.netty.channel.ChannelHandlerContext;
import io.netty.channel.SimpleChannelInboundHandler;
import io.netty.handler.codec.http.*;
import io.netty.handler.codec.http.websocketx.WebSocketFrame;
import io.netty.handler.codec.http.websocketx.WebSocketServerHandshaker;
import io.netty.handler.codec.http.websocketx.WebSocketServerHandshakerFactory;
import io.netty.util.CharsetUtil;
import io.netty.util.concurrent.EventExecutorGroup;
import io.netty.util.concurrent.Future;

import java.util.Arrays;
import java.util.concurrent.*;

import static io.netty.handler.codec.http.HttpHeaders.Names.HOST;
import static com.iotracks.ws.manager.WebSocketManager.OPCODE_MSG;

/**
 * @author iotracks
 */
public class RestApiHandler extends SimpleChannelInboundHandler {

    private final EventExecutorGroup executor;
    private final boolean ssl;

    public RestApiHandler(boolean ssl, EventExecutorGroup executor) {
        super(false);
        this.ssl = ssl;
        this.executor = executor;
    }

    @Override
    public void channelRead0(ChannelHandlerContext ctx, Object msg){
        if (msg instanceof FullHttpRequest) {
            handleHttpRequest(ctx, (FullHttpRequest) msg);
        } else if (msg instanceof WebSocketFrame) {
            handleWebSocketFrame(ctx, (WebSocketFrame) msg);
        }
    }

    private void handleWebSocketFrame(ChannelHandlerContext ctx, WebSocketFrame frame) {
        System.out.println("RestApiHandler.handleWebSocketFrame");
        ByteBuf content = frame.content();
        if(content.isReadable()) {
            Byte opcode = content.readByte();
            if (opcode == OPCODE_MSG.intValue()) {
                byte[] byteArray = new byte[content.readableBytes()];
                int readerIndex = content.readerIndex();
                content.getBytes(readerIndex, byteArray);
                int totalMsgLength = ByteUtils.bytesToInteger(Arrays.copyOfRange(byteArray, 0, 4));
                IOMessage message = new IOMessage(Arrays.copyOfRange(byteArray, 4, totalMsgLength + 4));
                System.out.println("Message: \n" + message.toString());
                LogStorage.addLog(new LogMessage(message));
            }
        }
    }

    private void handleHttpRequest(ChannelHandlerContext ctx, FullHttpRequest request) {
        System.out.println("RestApiHandler.handleHttpRequest");
        RestUrlType urlType = RestUrlType.getByUrl(request.getUri());
        if (urlType != null){
            runTask(new HttpRequestHandler(request, ctx.alloc().buffer(), urlType), ctx, request);
        } else {
            String prefix;
            if(ssl) {
                prefix = "wss://";
            } else {
                prefix = "ws://";
            }
            WebSocketServerHandshakerFactory wsFactory = new WebSocketServerHandshakerFactory(prefix + request.headers().get(HOST) + request.getUri(), null, true);
            WebSocketServerHandshaker handshaker = wsFactory.newHandshaker(request);
            if (handshaker == null) {
                WebSocketServerHandshakerFactory.sendUnsupportedVersionResponse(ctx.channel());
            } else {
                handshaker.handshake(ctx.channel(), request);
            }
            /*String errorMsg = "URL not supported: " + request.getUri();
            ByteBuf	errorMsgBytes = ctx.alloc().buffer();
            errorMsgBytes.writeBytes(errorMsg.getBytes());
            sendHttpResponse(ctx, request, new DefaultFullHttpResponse(HttpVersion.HTTP_1_1, HttpResponseStatus.NOT_FOUND, errorMsgBytes));*/
        }
    }

    private void runTask(Callable<? extends Object> callable, ChannelHandlerContext ctx, FullHttpRequest req) {
        final Future<? extends Object> future = executor.submit(callable);
        future.addListener((Future<Object> futureR) -> {
            if (futureR.isSuccess()) {
                try {
                    sendHttpResponse(ctx, req, (FullHttpResponse)futureR.get());
                } catch (InterruptedException | ExecutionException e){
                    ctx.fireExceptionCaught(e);
                    ctx.close();
                }
            } else {
                ctx.fireExceptionCaught(futureR.cause());
                ctx.close();
            }
        });
    }

    private static void sendHttpResponse( ChannelHandlerContext ctx, FullHttpRequest req, FullHttpResponse res){
        if (res.getStatus().code() != 200) {
            ByteBuf buf = Unpooled.copiedBuffer(res.getStatus().toString(), CharsetUtil.UTF_8);
            res.content().writeBytes(buf);
            buf.release();
            HttpHeaders.setContentLength(res, res.content().readableBytes());
        }

        ChannelFuture f = ctx.channel().writeAndFlush(res);
        if (!HttpHeaders.isKeepAlive(req) || res.getStatus().code() != 200) {
            f.addListener(ChannelFutureListener.CLOSE);
        }
    }

    @Override
    public void exceptionCaught(ChannelHandlerContext ctx, Throwable cause) throws Exception {
        super.exceptionCaught(ctx, cause);
    }


}
