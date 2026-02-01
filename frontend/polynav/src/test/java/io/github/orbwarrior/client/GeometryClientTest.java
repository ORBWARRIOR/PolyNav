package io.github.orbwarrior.client;

import fyp.generated.GeometryServiceGrpc;
import fyp.generated.MapData;
import fyp.generated.Point;
import fyp.generated.SaveMapResponse;
import fyp.generated.Triangle;
import fyp.generated.TriangulationResult;
import io.grpc.ManagedChannel;
import io.grpc.inprocess.InProcessChannelBuilder;
import io.grpc.inprocess.InProcessServerBuilder;
import io.grpc.stub.StreamObserver;
import io.grpc.testing. GrpcCleanupRule;
import javafx.geometry.Point2D;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

import java.util.Arrays;
import java.util.List;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.verify;

public class GeometryClientTest {

    private String serverName;
    private ManagedChannel channel;
    private GeometryClient client;
    private GeometryServiceGrpc.GeometryServiceImplBase serviceImpl;

    @BeforeEach
    public void setUp() throws Exception {
        // Generate a unique in-process server name.
        serverName = InProcessServerBuilder.generateName();

        // Create a mock implementation of the service.
        serviceImpl = new GeometryServiceGrpc.GeometryServiceImplBase() {
            @Override
            public void triangulate(MapData request, StreamObserver<TriangulationResult> responseObserver) {
                // Return a dummy result
                Triangle t1 = Triangle.newBuilder()
                        .setA(Point.newBuilder().setX(0).setY(0).build())
                        .setB(Point.newBuilder().setX(10).setY(0).build())
                        .setC(Point.newBuilder().setX(0).setY(10).build())
                        .build();
                
                responseObserver.onNext(TriangulationResult.newBuilder()
                        .addTriangles(t1)
                        .build());
                responseObserver.onCompleted();
            }
        };

        // Create a server, add service, start, and register for automatic graceful shutdown.
        InProcessServerBuilder.forName(serverName).directExecutor().addService(serviceImpl).build().start();

        // Create a client channel and register for automatic graceful shutdown.
        channel = InProcessChannelBuilder.forName(serverName).directExecutor().build();

        // Create the client
        client = new GeometryClient(channel);
    }

    @AfterEach
    public void tearDown() throws Exception {
        client.shutdown();
    }

    @Test
    public void testGetTriangulation() {
        List<Point2D> inputPoints = Arrays.asList(
                new Point2D(0, 0),
                new Point2D(10, 0),
                new Point2D(0, 10)
        );

        List<Triangle> result = client.getTriangulation(inputPoints);

        assertNotNull(result);
        assertEquals(1, result.size());
        Triangle t = result.get(0);
        assertEquals(0, t.getA().getX(), 0.001);
        assertEquals(0, t.getA().getY(), 0.001);
        assertEquals(10, t.getB().getX(), 0.001);
    }
}