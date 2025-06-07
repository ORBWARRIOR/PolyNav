package fyp;

import static org.junit.jupiter.api.Assertions.assertEquals;

import java.util.ArrayList;
import java.util.List;

import org.junit.jupiter.api.Test;

class DelaunayTriangulationTest {

    public void setup() {
        
    }

    public void tearDown() {

    }
    
    @Test
    public void testTriangulate() {
        // Setup a point cloud for testing
        List<Point> pointCloud = new ArrayList<>();
        pointCloud.add(new Point(0, 7));
        pointCloud.add(new Point(-5, 5));
        pointCloud.add(new Point(5, 5));
        pointCloud.add(new Point(-2, 3));
        pointCloud.add(new Point(3, 1));
        pointCloud.add(new Point(-4, -1));
        pointCloud.add(new Point(1, -2));
        pointCloud.add(new Point(-6, -4));
        pointCloud.add(new Point(5, -4));

        // Create an instance of DelaunayTriangulation
        DelaunayTriangulation dt = new DelaunayTriangulation(pointCloud);
        // Call the triangulate method
        dt.triangulate();
    }
}
