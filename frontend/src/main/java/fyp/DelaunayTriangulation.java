package fyp;

import java.util.List;
import java.util.ArrayList;
import java.util.Stack;

/**
 * Implements the Delaunay triangulation algorithm for a given point cloud. It
 * sorts the points into bins and processes them in snake-like pattern to
 * create triangles while maintaining the Delaunay condition.
 */
public class DelaunayTriangulation {

    private List<Point> pointCloud;
    private Stack<Triangle> triangles;
    private List<Point> workingPointCloud;

    /**
     * Constructor for the DelaunayTriangulation class
     * 
     * @param pointCloud the list of points to be triangulated
     */
    public DelaunayTriangulation(List<Point> pointCloud) {
        this.pointCloud = pointCloud;
    }

    /**
     * Function to call the triangulation algorithm
     */
    public void triangulate() {
        double[] remapFactors = remapToUnitSq();
        constructTriangulation();
        remapToOriginalSize(remapFactors);
    }

    private void constructTriangulation() {
        List<List<Point>> bins = binSort(pointCloud);
        makeSuperTriangle();
        for (int i = 0; i < bins.size(); i++) {
            List<Point> bin = bins.get(i);
            for (int j = 0; j < bin.size(); j++) {
                workingPointCloud.add(bin.get(i));
                locateWhichTrianglePointIsIn();
                makeTriangles();
                if (!checkDelaunayProperty()) {
                    maintainDelaunay();
                }
            }
        }
    }

    private List<List<Point>> binSort(List<Point> pointCloud) {
        int gridLength = determineGridLength();
        double binSize = 1.0 / gridLength;
        List<List<List<Point>>> bins = initiliseBins(gridLength);
        assignPointsToBins(bins, gridLength, binSize);
        return flattenBins(bins, gridLength);
    }

    private int determineGridLength() {
        int n = (int) Math.sqrt(pointCloud.size());
        if (n % 2 == 1) {
            n++; // Ensure n is even
        }
        return n;
    }

    private List<List<List<Point>>> initiliseBins(int gridLength) {
        List<List<List<Point>>> bins = new ArrayList<>();
        for (int y = 0; y < gridLength; y++) {
            List<List<Point>> row = new ArrayList<>();
            for (int x = 0; x < gridLength; x++) {
                row.add(new ArrayList<>());
            }
            bins.add(row);
        }
        return bins;
    }

    private void assignPointsToBins( List<List<List<Point>>> bins, int gridLength, double binSize) {
        List<Point> temporaryPointCloud = new ArrayList<>(initialSort());
        for (Point point : temporaryPointCloud) {
            // Calculate x and y bins
            int xBin = (int) (point.x / binSize);
            int yBin = gridLength - 1 - (int) (point.y / binSize);
            xBin = Math.min(xBin, gridLength - 1);
            yBin = Math.min(yBin, gridLength - 1);

            // Reverse x direction for odd rows (snake pattern)
            if (yBin % 2 == 1) {
                xBin = gridLength - 1 - xBin;
            }
            bins.get(yBin).get(xBin).add(point);
        }
    }

    private List<Point> initialSort() {
        // Sort by ascending y (bottom to top) and descending x (right to left)
        pointCloud.sort((p1, p2) -> {
            int cmpY = Double.compare(p1.y, p2.y); // Ascending y (prioritize lower y = bottom rows)
            if (cmpY != 0)
                return cmpY;
            return Double.compare(p2.x, p1.x); // Descending x (rightmost first)
        });
        return pointCloud;
    }

    private List<List<Point>> flattenBins(List<List<List<Point>>> bins, int gridLength) {
        List<List<Point>> flattenedBins = new ArrayList<>();
        for (int y = 0; y < gridLength; y++) {
            List<List<Point>> row = bins.get(y);
            if (y % 2 == 0) {
                flattenedBins.addAll(row); // Even row: left-to-right
            } else {
                // Reverse the row for odd rows
                for (int x = row.size() - 1; x >= 0; x--) {
                    flattenedBins.add(row.get(x));
                }
            }
        }
        return flattenedBins;
    }

    private double[] remapToUnitSq() {
        double minX = Double.MAX_VALUE;
        double minY = Double.MAX_VALUE;
        double maxX = Double.MIN_VALUE;
        double maxY = Double.MIN_VALUE;

        // Find the min and max x and y values
        for (int i = 0; i < pointCloud.size(); i++) {
            double x = pointCloud.get(i).x;
            double y = pointCloud.get(i).y;
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

        // Set the scale to the largest range (min is 1)
        double scale = maxX - minX;
        if (maxY - minY > scale) {
            scale = maxY - minY;
        }
        if (scale < 1) {
            scale = 1;
        }

        // Remap the points to a unit square
        for (int i = 0; i < pointCloud.size(); i++) {
            pointCloud.get(i).x = (pointCloud.get(i).x - minX) / scale;
            pointCloud.get(i).y = (pointCloud.get(i).y - minY) / scale;
        }
        double[] remapFactors = new double[3];
        remapFactors[0] = scale;
        remapFactors[1] = minX;
        remapFactors[2] = minY;
        return remapFactors;
    }

    private void remapToOriginalSize(double[] remapFactors) {
        for (int i = 0; i < workingPointCloud.size(); i++) {
            workingPointCloud.get(i).x = (workingPointCloud.get(i).x * remapFactors[0]) + remapFactors[1];
            workingPointCloud.get(i).y = (workingPointCloud.get(i).y * remapFactors[0]) + remapFactors[2];
        }
    }

    private void makeSuperTriangle() {
        Point p1 = new Point(0, 25);
        Point p2 = new Point(-20, -15);
        Point p3 = new Point(20, -15);
        triangles.push(new Triangle(p1, p2, p3));
        workingPointCloud = new ArrayList<>();
        workingPointCloud.add(p1);
        workingPointCloud.add(p2);
        workingPointCloud.add(p3);
    }

    private void locateWhichTrianglePointIsIn() {
        Point point = workingPointCloud.getLast();
        Triangle triangleThatPointIsIn = null;
        //will always empty the stack except for the super triangle
        while (triangles.size() > 1) {
            Triangle triangle = triangles.pop();
            if (triangleThatPointIsIn == null && triangle.containsPoint(point)) {
                triangleThatPointIsIn = triangle;
            }
        }
        if (triangleThatPointIsIn != null) {
            triangles.push(triangleThatPointIsIn);
        }
    }

    private void makeTriangles() {
        Triangle triangle = triangles.pop();
        
    }

    private boolean checkDelaunayProperty() {
        return true;
    }

    private void maintainDelaunay() {

    }
}
