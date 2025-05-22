package fyp;

/**
 * Triangle class to represent a triangle
 */
public class Triangle {
    private Point p1;
    private Point p2;
    private Point p3;
    private Edge e1;
    private Edge e2;
    private Edge e3;

    public Triangle(Point p1, Point p2, Point p3) {
        this.p1 = p1;
        this.p2 = p2;
        this.p3 = p3;
        e1 = new Edge(p1, p2);
        e2 = new Edge(p2, p3);
        e3 = new Edge(p3, p1);
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
    public Edge getE1() {
        return e1;
    }
    public Edge getE2() {
        return e2;
    }
    public Edge getE3() {
        return e3;
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
    public void setE1(Edge e1) {
        this.e1 = e1;
    }
    public void setE2(Edge e2) {
        this.e2 = e2;
    }
    public void setE3(Edge e3) {
        this.e3 = e3;
    }
}
