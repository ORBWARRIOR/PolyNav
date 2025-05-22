package fyp;

import java.util.List;
import java.util.ArrayList;
import java.util.Queue;

public class DelaunayTriangulation {

    private List<Point> mapOfPoints;
    private Queue<Triangle> triangles;

    public DelaunayTriangulation(List<Point> mapOfPoints) {
        this.mapOfPoints = mapOfPoints; 
    }

    /**
     * Function to call the triangulation algorithm
     */
    public void triangulate() {
        double[] remapFactors = remapToUnitSq();
        binSort(mapOfPoints);
        List<Point> workingMapOfPoints = new ArrayList<>(makeSuperTriangle());
        
        for (int i = 0; i < mapOfPoints.size(); i++) {
            workingMapOfPoints.add(mapOfPoints.get(i));
            if (checkDelaunay(workingMapOfPoints)) {
                makeTriangles(workingMapOfPoints);
            } else {
                maintainDelaunay(workingMapOfPoints);
                
            }
        }



        remapToOriginalSize(remapFactors);
    }

    //TO DO
    private void binSort(List<Point> mapOfPoints) {
        int n = (int) Math.sqrt(mapOfPoints.size());
        double binSize = 1.0 / n;
        List<List<Point>> bins = new ArrayList<>(n);
    }
    
    /**
     * Function to find the scale of the points and remap them to a unit square
     * 
     * @return double[] array containing the scale factor and translation factor
     */
    private double[] remapToUnitSq() {
        double minX = Double.MAX_VALUE;
        double minY = Double.MAX_VALUE;
        double maxX = Double.MIN_VALUE;
        double maxY = Double.MIN_VALUE;

        // Find the min and max x and y values
        for (int i = 0; i < mapOfPoints.size(); i++) {
            double x = mapOfPoints.get(i).x;
            double y = mapOfPoints.get(i).y;
            if (x < minX) {
                minX = x;
            } else if (x > maxX) {
                maxX = x;
            }
            if (y < minY) {
                minY = y;
            } else if (y > maxY) {
                maxY = y;
            }
        }

        //Set the scale to the largest range (min is 1)
        double scale = maxX - minX;
        if (maxY - minY > scale) {
            scale = maxY - minY;
        }
        if (scale < 1) {
            scale = 1;
        }

        // Remap the points to a unit square
        for (int i = 0; i < mapOfPoints.size(); i++) {
            mapOfPoints.get(i).x = (mapOfPoints.get(i).x - minX) / scale;
            mapOfPoints.get(i).y = (mapOfPoints.get(i).y - minY) / scale;
        }
        double[] remapFactors = new double[3];
        remapFactors[0] = scale;
        remapFactors[1] = minX;
        remapFactors[2] = minY;
        return remapFactors;
    }

    /**
     * Remap the points back to their original size
     * @param mapOfPoints the list of points to be remapped
     * @param remapFactors the scale and translation factors
     */
    private void remapToOriginalSize(double[] remapFactors) {
        for (int i = 0; i < mapOfPoints.size(); i++) {
            mapOfPoints.get(i).x = (mapOfPoints.get(i).x * remapFactors[0]) + remapFactors[1];
            mapOfPoints.get(i).y = (mapOfPoints.get(i).y * remapFactors[0]) + remapFactors[2];
        }
    }

    public List<Point> makeSuperTriangle() {
        Point p1 = new Point(0, 10);
        Point p2 = new Point(-10, -10);
        Point p3 = new Point(10, 10);
        Triangle superTriangle = new Triangle(p1, p2, p3);
        triangles.add(superTriangle);
        List<Point> workingMapOfPoints = new ArrayList<>();
        workingMapOfPoints.add(p1);
        workingMapOfPoints.add(p2);
        workingMapOfPoints.add(p3);
        return workingMapOfPoints;
    }

    //TO DO
    private boolean checkDelaunay(List<Point> workingMapOfPoints) {
        // Check if the points are in the circumcircle of the previous triangles
        for (int i = 0; i < triangles.size(); i++) {
            Triangle triangle = triangles.poll();
            Point p1 = triangle.getP1();
            Point p2 = triangle.getP2();
            Point p3 = triangle.getP3();

            if (distanceToCircumcenter < circumRadius) {
                return false;
            }
        }
        // Check if the points are in the circumcircle of the super triangle
        return true;
    }

    private void makeTriangles(List<Point> workingMapOfPoints) {
        
    }

    private void maintainDelaunay(List<Point> workingMapOfPoints) {
        
    }
}
