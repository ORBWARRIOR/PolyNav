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
        this(ManagedChannelBuilder.forAddress(host, port)
                .usePlaintext()
                .build());
    }

    public GeometryClient(ManagedChannel channel) {
        this.channel = channel;
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

        return executeRequest(request);
    }

    public List<Triangle> getTriangulation(List<Point2D> points, List<int[]> segments) {
        MapData.Builder requestBuilder = MapData.newBuilder();

        // Convert each segment into a 2-point Obstacle
        for (int[] seg : segments) {
            if (seg.length < 2) continue;
            Point p1 = Point.newBuilder().setX(points.get(seg[0]).getX()).setY(points.get(seg[0]).getY()).build();
            Point p2 = Point.newBuilder().setX(points.get(seg[1]).getX()).setY(points.get(seg[1]).getY()).build();
            
            // Hack: Send p1-p2-p1 to bypass "len < 3" check in backend and enforce constraint p1-p2
            requestBuilder.addObstacles(Obstacle.newBuilder().addPoints(p1).addPoints(p2).addPoints(p1).build());
        }

        return executeRequest(requestBuilder.build());
    }

    private List<Triangle> executeRequest(MapData request) {
        try {
            TriangulationResult result = blockingStub.triangulate(request);
            return result.getTrianglesList();
        } catch (Exception e) {
            System.err.println("RPC failed: " + e.getMessage());
            return new ArrayList<>();
        }
    }
}
