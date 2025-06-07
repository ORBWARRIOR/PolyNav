package fyp;

/**
 * Triangle class to represent a triangle
 */
public class Triangle {
    private Point p1;
    private Point p2;
    private Point p3;

    public Triangle(Point p1, Point p2, Point p3) {
        this.p1 = p1;
        this.p2 = p2;
        this.p3 = p3;
        linkPoints();
    }

    private void linkPoints() {
        p1.addNeighbour(p2);
        p1.addNeighbour(p3);
        p2.addNeighbour(p1);
        p2.addNeighbour(p3);
        p3.addNeighbour(p1);
        p3.addNeighbour(p2);
        p1.addTriangle(this);
        p2.addTriangle(this);
        p3.addTriangle(this);
    }

    public Point getP1() {
        return p1;
    }
    public Point getP2() {
        return p2;
    }
    public Point getP3() {
        return p3;
    }
    public void setP1(Point p1) {
        this.p1 = p1;
    }
    public void setP2(Point p2) {
        this.p2 = p2;
    }
    public void setP3(Point p3) {
        this.p3 = p3;
    }

    /**
     * Uses the barycentric coordinates to determine if a point is inside the triangle
     * @param point the point to check
     * @return true if the point is inside the triangle, false otherwise
     */
    public boolean containsPoint(Point point) {
        double denominator = ((p2.y - p3.y) * (p1.x - p3.x) + (p3.x - p2.x) * (p1.y - p3.y));
        if (denominator == 0) {
            return false; // Degenerate triangle, cannot contain points
        }
        double a = ((p2.y - p3.y) * (point.x - p3.x) + (p3.x - p2.x) * (point.y - p3.y)) / denominator;
        double b = ((p3.y - p1.y) * (point.x - p3.x) + (p1.x - p3.x) * (point.y - p3.y)) / denominator;
        double c = 1 - a - b;
        return (a >= 0 && b >= 0 && c >= 0);
    }
}
