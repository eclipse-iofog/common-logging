package com.iotracks.rest;

import io.netty.bootstrap.ServerBootstrap;
import io.netty.channel.Channel;
import io.netty.channel.ChannelInitializer;
import io.netty.channel.ChannelPipeline;
import io.netty.channel.EventLoopGroup;
import io.netty.channel.nio.NioEventLoopGroup;
import io.netty.channel.socket.nio.NioServerSocketChannel;
import io.netty.handler.codec.http.HttpObjectAggregator;
import io.netty.handler.codec.http.HttpServerCodec;
import io.netty.handler.ssl.SslContext;
import io.netty.handler.ssl.SslContextBuilder;
import io.netty.handler.ssl.util.SelfSignedCertificate;
import io.netty.util.concurrent.DefaultEventExecutorGroup;
import io.netty.util.concurrent.EventExecutorGroup;

import javax.net.ssl.SSLException;
import java.security.cert.CertificateException;

public class RestApiServer implements Runnable {

    private final EventLoopGroup bossGroup = new NioEventLoopGroup(1);
    private final EventLoopGroup workerGroup = new NioEventLoopGroup(10);
    private final EventExecutorGroup executor = new DefaultEventExecutorGroup(10);

    static final boolean SSL = System.getProperty("ssl") != null;
    static final int PORT = 10555;

    public void run() {
        try {
            final SslContext sslCtx;
            if (SSL) {
                SelfSignedCertificate ssc = new SelfSignedCertificate();
                sslCtx = SslContextBuilder.forServer(ssc.certificate(), ssc.privateKey()).build();
            } else {
                sslCtx = null;
            }
            try {
                ServerBootstrap b = new ServerBootstrap();
                b.group(bossGroup, workerGroup)
                        .channel(NioServerSocketChannel.class)
                        .childHandler(new ChannelInitializer() {
                            @Override
                            protected void initChannel(Channel channel) throws Exception {
                                ChannelPipeline pipeline = channel.pipeline();
                                if (sslCtx != null) {
                                    pipeline.addLast(sslCtx.newHandler(channel.alloc()));
                                }
                                pipeline.addLast(new HttpServerCodec());
                                pipeline.addLast(new HttpObjectAggregator(Integer.MAX_VALUE));
                                pipeline.addLast(new RestApiHandler(sslCtx != null, executor));
                            }
                        });
                Channel ch = b.bind(PORT).sync().channel();
                System.out.println("Rest api server started at port: " + PORT);
                ch.closeFuture().sync();
            } finally {
                bossGroup.shutdownGracefully();
                workerGroup.shutdownGracefully();
            }
        } catch (CertificateException ce) {
            System.err.println("Rest api server CertificateException : " + ce.getMessage());
        } catch (SSLException ssle) {
            System.err.println("Rest api server SSLException : " + ssle.getMessage());
        } catch (InterruptedException ie) {
            System.err.println("Rest api server InterruptedException : " + ie.getMessage());
        }
    }

    public void stop() throws Exception {
        bossGroup.shutdownGracefully();
        workerGroup.shutdownGracefully();
    }
}
