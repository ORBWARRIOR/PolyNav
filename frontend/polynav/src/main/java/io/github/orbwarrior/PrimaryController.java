package io.github.orbwarrior;

import com.google.gson.Gson;
import com.google.gson.JsonArray;
import com.google.gson.JsonElement;
import com.google.gson.JsonParser;
import com.google.gson.reflect.TypeToken;
import fyp.generated.Triangle;
import io.github.orbwarrior.client.GeometryClient;
import javafx.fxml.FXML;
import javafx.geometry.Point2D;
import javafx.scene.canvas.Canvas;
import javafx.scene.canvas.GraphicsContext;
import javafx.scene.control.Alert;
import javafx.scene.control.CheckBox;
import javafx.scene.input.MouseEvent;
import javafx.scene.layout.BorderPane;
import javafx.scene.layout.Pane;
import javafx.scene.paint.Color;
import javafx.stage.FileChooser;
import java.io.File;
import java.io.FileReader;
import java.io.IOException;
import java.lang.reflect.Type;
import java.util.ArrayList;
import java.util.HashSet;
import java.util.List;
import java.util.Set;

public class PrimaryController {

    @FXML
    private BorderPane mainContent;
    @FXML
    private Pane canvasContainer;
    @FXML
    private Canvas drawingCanvas;
    @FXML
    private CheckBox liveModeCheckbox;

    private final List<Point2D> points = new ArrayList<>();
    private GeometryClient client;
    
    // Viewport transform state
    private double scale = 1.0;
    private double offsetX = 0.0;
    private double offsetY = 0.0;

    // Helper class for JSON import
    public static class JsonPoint {
        public double x, y;
    }
    
    // Helper for Map Import
    public static class MapFile {
        public List<JsonPoint> points;
        public List<int[]> segments;
        public List<int[]> constraints; // Alias for segments
    }

    // Helper class for Debug JSON import
    public static class DebugTriangle {
        public int id;
        public List<JsonPoint> points;
        public List<Integer> neighbours;
    }

    public void setClient(GeometryClient client) {
        this.client = client;
    }

    @FXML
    public void initialise() {
        // Resize canvas when window resizes
        canvasContainer.widthProperty().addListener((obs, oldVal, newVal) -> {
            drawingCanvas.setWidth(newVal.doubleValue());
            recalcBoundsAndRedraw();
        });
        canvasContainer.heightProperty().addListener((obs, oldVal, newVal) -> {
            drawingCanvas.setHeight(newVal.doubleValue());
            recalcBoundsAndRedraw();
        });
    }

    @FXML
    private void showDelaunayView() {
        // Already in Delaunay view
    }

    @FXML
    private void showPathPlanningView() {
        Alert alert = new Alert(Alert.AlertType.INFORMATION);
        alert.setTitle("Not Implemented");
        alert.setHeaderText(null);
        alert.setContentText("Path Planning algorithm is not yet implemented.");
        alert.showAndWait();
    }

    @FXML
    private void handleCanvasClick(MouseEvent event) {
        // Inverse transform to get world coordinates
        double worldX = (event.getX() - offsetX) / scale;
        double worldY = (event.getY() - offsetY) / scale;
        
        points.add(new Point2D(worldX, worldY));
        
        // Don't auto-zoom on click adding, just redraw
        redraw();
        
        if (liveModeCheckbox.isSelected()) {
            handleTriangulate();
        }
    }

    @FXML
    private void handleClear() {
        points.clear();
        scale = 1.0;
        offsetX = 0.0;
        offsetY = 0.0;
        redraw();
    }

    @FXML
    private void handleTriangulate() {
        if (client == null || points.size() < 3) return;

        List<Triangle> triangles = client.getTriangulation(points);
        drawTriangles(triangles);
    }

    @FXML
    private void handleImport() {
        FileChooser fileChooser = new FileChooser();
        fileChooser.setTitle("Import Map JSON");
        fileChooser.getExtensionFilters().add(new FileChooser.ExtensionFilter("JSON Files", "*.json"));
        File selectedFile = fileChooser.showOpenDialog(drawingCanvas.getScene().getWindow());

        if (selectedFile != null) {
            try (FileReader reader = new FileReader(selectedFile)) {
                JsonElement root = JsonParser.parseReader(reader);
                
                if (root.isJsonObject()) {
                    // New Map Format with "points" and "segments"
                    handleMapImport(root);
                } else if (root.isJsonArray()) {
                    JsonArray array = root.getAsJsonArray();
                    if (array.size() > 0) {
                        JsonElement firstItem = array.get(0);
                        if (firstItem.isJsonObject() && firstItem.getAsJsonObject().has("points") && firstItem.getAsJsonObject().get("points").isJsonArray()) {
                            handleDebugJsonImport(root);
                        } else {
                            handleSimpleJsonImport(root);
                        }
                    }
                }
            } catch (Exception e) {
                Alert alert = new Alert(Alert.AlertType.ERROR);
                alert.setTitle("Import Error");
                alert.setHeaderText("Failed to import map");
                alert.setContentText(e.getMessage());
                alert.showAndWait();
            }
        }
    }

    private void handleMapImport(JsonElement root) {
        Gson gson = new Gson();
        MapFile map = gson.fromJson(root, MapFile.class);
        
        if (map != null && map.points != null) {
            points.clear();
            for (JsonPoint p : map.points) {
                points.add(new Point2D(p.x, p.y));
            }
            
            // If we have explicit segments, use them. Otherwise just points.
            List<int[]> segs = map.segments != null ? map.segments : map.constraints;
            
            if (segs != null && !segs.isEmpty() && client != null) {
                // If live mode is on or just for loading, we might want to store segments.
                // For now, we immediately triangulate with segments if available.
                recalcBoundsAndRedraw();
                if (liveModeCheckbox.isSelected()) {
                    List<Triangle> triangles = client.getTriangulation(points, segs);
                    drawTriangles(triangles);
                }
            } else {
                finishImport();
            }
        }
    }

    private void handleSimpleJsonImport(JsonElement root) {
        Gson gson = new Gson();
        Type listType = new TypeToken<ArrayList<JsonPoint>>(){}.getType();
        List<JsonPoint> importedPoints = gson.fromJson(root, listType);

        if (importedPoints != null) {
            points.clear();
            for (JsonPoint p : importedPoints) {
                points.add(new Point2D(p.x, p.y));
            }
            finishImport();
        }
    }

    private void handleDebugJsonImport(JsonElement root) {
        Gson gson = new Gson();
        Type listType = new TypeToken<ArrayList<DebugTriangle>>(){}.getType();
        List<DebugTriangle> triangles = gson.fromJson(root, listType);

        if (triangles != null) {
            Set<Point2D> uniquePoints = new HashSet<>();
            for (DebugTriangle t : triangles) {
                if (t.points != null) {
                    for (JsonPoint p : t.points) {
                        uniquePoints.add(new Point2D(p.x, p.y));
                    }
                }
            }
            points.clear();
            points.addAll(uniquePoints);
            finishImport();
        }
    }

    private void finishImport() {
        recalcBoundsAndRedraw();
        if (liveModeCheckbox.isSelected()) {
            handleTriangulate();
        }
    }

    private void recalcBoundsAndRedraw() {
        if (points.isEmpty()) {
            scale = 1.0;
            offsetX = drawingCanvas.getWidth() / 2;
            offsetY = drawingCanvas.getHeight() / 2;
            redraw();
            return;
        }

        double minX = Double.MAX_VALUE;
        double maxX = -Double.MAX_VALUE;
        double minY = Double.MAX_VALUE;
        double maxY = -Double.MAX_VALUE;

        for (Point2D p : points) {
            if (p.getX() < minX) minX = p.getX();
            if (p.getX() > maxX) maxX = p.getX();
            if (p.getY() < minY) minY = p.getY();
            if (p.getY() > maxY) maxY = p.getY();
        }

        double width = maxX - minX;
        double height = maxY - minY;
        
        // Handle single point or very small bounds
        if (width < 1e-9) width = 1.0;
        if (height < 1e-9) height = 1.0;

        double canvasWidth = drawingCanvas.getWidth();
        double canvasHeight = drawingCanvas.getHeight();
        
        // Fit to 90% of screen
        double scaleX = (canvasWidth * 0.9) / width;
        double scaleY = (canvasHeight * 0.9) / height;
        scale = Math.min(scaleX, scaleY);

        double centerX = minX + width / 2.0;
        double centerY = minY + height / 2.0;

        offsetX = (canvasWidth / 2.0) - (centerX * scale);
        offsetY = (canvasHeight / 2.0) - (centerY * scale);

        redraw();
    }

    private void redraw() {
        GraphicsContext gc = drawingCanvas.getGraphicsContext2D();
        gc.clearRect(0, 0, drawingCanvas.getWidth(), drawingCanvas.getHeight());

        // Draw axes (optional, but helpful)
        gc.setStroke(Color.LIGHTGRAY);
        gc.setLineWidth(0.5);
        // X-axis
        double yAxisY = 0 * scale + offsetY;
        gc.strokeLine(0, yAxisY, drawingCanvas.getWidth(), yAxisY);
        // Y-axis
        double xAxisX = 0 * scale + offsetX;
        gc.strokeLine(xAxisX, 0, xAxisX, drawingCanvas.getHeight());

        gc.setFill(Color.BLUE);
        for (Point2D p : points) {
            double screenX = p.getX() * scale + offsetX;
            double screenY = p.getY() * scale + offsetY;
            gc.fillOval(screenX - 3, screenY - 3, 6, 6);
        }
    }

    private void drawTriangles(List<Triangle> triangles) {
        redraw(); // Clear and draw points first
        GraphicsContext gc = drawingCanvas.getGraphicsContext2D();
        gc.setStroke(Color.RED);
        gc.setLineWidth(1.0);

        for (Triangle t : triangles) {
            double[] xPoints = {
                t.getA().getX() * scale + offsetX, 
                t.getB().getX() * scale + offsetX, 
                t.getC().getX() * scale + offsetX
            };
            double[] yPoints = {
                t.getA().getY() * scale + offsetY, 
                t.getB().getY() * scale + offsetY, 
                t.getC().getY() * scale + offsetY
            };
            gc.strokePolygon(xPoints, yPoints, 3);
        }
    }
}
