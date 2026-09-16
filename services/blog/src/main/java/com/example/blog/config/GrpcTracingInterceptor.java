package com.example.blog.config;

import io.grpc.Metadata;
import io.grpc.ServerCall;
import io.grpc.ServerCallHandler;
import io.grpc.ServerInterceptor;
import io.grpc.ForwardingServerCallListener;
import io.micrometer.tracing.Span;
import io.micrometer.tracing.Tracer;
import io.micrometer.tracing.propagation.Propagator;
import lombok.RequiredArgsConstructor;
import net.devh.boot.grpc.server.interceptor.GrpcGlobalServerInterceptor;

/**
 * Extracts an inbound W3C traceparent (injected by the gateway's otelgrpc
 * client interceptor, or by another Go/Node service dialing blog directly)
 * from gRPC metadata and starts a child span, so a trace entering blog via
 * gRPC continues the caller's trace instead of starting a new one.
 * <p>
 * HTTP requests already get this for free from Spring Boot's own web
 * tracing autoconfiguration; net.devh's gRPC server has no such
 * autoconfiguration, so it's done by hand here using the same
 * Tracer/Propagator beans (Micrometer Tracing, OTel bridge) the rest of
 * the app already uses.
 */
@GrpcGlobalServerInterceptor
@RequiredArgsConstructor
public class GrpcTracingInterceptor implements ServerInterceptor {

    private static final Propagator.Getter<Metadata> GETTER = (carrier, key) -> {
        Metadata.Key<String> metadataKey = Metadata.Key.of(key, Metadata.ASCII_STRING_MARSHALLER);
        return carrier.get(metadataKey);
    };

    private final Tracer tracer;
    private final Propagator propagator;

    @Override
    public <ReqT, RespT> ServerCall.Listener<ReqT> interceptCall(
            ServerCall<ReqT, RespT> call, Metadata headers, ServerCallHandler<ReqT, RespT> next) {

        Span span = propagator.extract(headers, GETTER)
                .name("grpc " + call.getMethodDescriptor().getFullMethodName())
                .kind(Span.Kind.SERVER)
                .start();

        ServerCall.Listener<ReqT> delegate;
        try (Tracer.SpanInScope ignored = tracer.withSpan(span)) {
            delegate = next.startCall(call, headers);
        }

        return new ForwardingServerCallListener.SimpleForwardingServerCallListener<>(delegate) {
            @Override
            public void onHalfClose() {
                try (Tracer.SpanInScope ignored = tracer.withSpan(span)) {
                    super.onHalfClose();
                } catch (RuntimeException | Error e) {
                    span.error(e);
                    throw e;
                }
            }

            @Override
            public void onCancel() {
                span.tag("grpc.canceled", "true");
                span.end();
                super.onCancel();
            }

            @Override
            public void onComplete() {
                span.end();
                super.onComplete();
            }
        };
    }
}
