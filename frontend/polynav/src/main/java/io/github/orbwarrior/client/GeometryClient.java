package io.github.orbwarrior.client;

import fyp.generated.GeometryServiceGrpc;
import fyp.generated.MapData;
import fyp.generated.Obstacle;
import fyp.generated.Point;
import fyp.generated.Triangle;
import fyp.generated.TriangulationResult;
import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;

import java.util.ArrayList;
import java.util.List;
import java.util.concurrent.TimeUnit;
import javafx.geometry.Point2D;

public class GeometryClient {

    private final ManagedChannel channel;
    private final GeometryServiceGrpc.GeometryServiceBlockingStub blockingStub;

    public GeometryClient(String host, int port) {
        this.channel = ManagedChannelBuilder.forAddress(host, port)
                .usePlaintext()
                .build();
        this.blockingStub = GeometryServiceGrpc.newBlockingStub(channel);
    }

    public void shutdown() throws InterruptedException {
        channel.shutdown().awaitTermination(5, TimeUnit.SECONDS);
    }

    public List<Triangle> getTriangulation(List<Point2D> points) {
        // Build the request
        Obstacle.Builder obstacleBuilder = Obstacle.newBuilder();
        for (Point2D p : points) {
            obstacleBuilder.addPoints(Point.newBuilder()
                    .setX(p.getX())
                    .setY(p.getY())
                    .build());
        }

        MapData request = MapData.newBuilder()
                .addObstacles(obstacleBuilder.build())
                .build();

        // Call the RPC
        try {
            TriangulationResult result = blockingStub.triangulate(request);
            return result.getTrianglesList();
        } catch (Exception e) {
            System.err.println("RPC failed: " + e.getMessage());
            return new ArrayList<>();
        }
    }
}
