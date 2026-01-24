package fyp;

import javafx.application.Application;
import javafx.fxml.FXMLLoader;
import javafx.scene.Parent;
import javafx.scene.Scene;
import javafx.stage.Stage;
import java.util.ArrayList;

import java.io.IOException;

/**
 * JavaFX App
 */
public class App extends Application {

    private static Scene scene;

    @Override
    public void start(Stage stage) throws IOException {
        scene = new Scene(loadFXML("primary"), 640, 480);
        stage.setScene(scene);
        stage.show();
    }

    static void setRoot(String fxml) throws IOException {
        scene.setRoot(loadFXML(fxml));
    }

    private static Parent loadFXML(String fxml) throws IOException {
        FXMLLoader fxmlLoader = new FXMLLoader(App.class.getResource(fxml + ".fxml"));
        return fxmlLoader.load();
    }

    public static void main(String[] args) {
        //launch();
        // Test 1
        ArrayList<Point> test1 = new ArrayList<>();
        test1.add(new Point(0, 7));
        test1.add(new Point(-5, 5));
        test1.add(new Point(5, 5));
        test1.add(new Point(-2, 3));
        test1.add(new Point(3, 1));
        test1.add(new Point(-4, -1));
        test1.add(new Point(1, -2));
        test1.add(new Point(-6, -4));
        test1.add(new Point(5, -4));

        DelaunayTriangulation dt = new DelaunayTriangulation(test1);
        dt.triangulate();

        // Test 2
        /*ArrayList<Point> test2 = new ArrayList<>();
        test2.add(new Point(1, 1));
        test2.add(new Point(3, 4));
        test2.add(new Point(-2, 3));
        test2.add(new Point(-2, 2));
        test2.add(new Point(-1, -1));
        test2.add(new Point(-2, -3));
        test2.add(new Point(4, -2));

        dt = new DelaunayTriangulation(test2);
        dt.triangulate();*/
    }

}