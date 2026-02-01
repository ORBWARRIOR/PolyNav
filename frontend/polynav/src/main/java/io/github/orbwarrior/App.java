package io.github.orbwarrior;

import io.github.orbwarrior.client.GeometryClient;
import javafx.application.Application;
import javafx.fxml.FXMLLoader;
import javafx.scene.Parent;
import javafx.scene.Scene;
import javafx.stage.Stage;
import java.io.IOException;

/**
 * JavaFX App
 */
public class App extends Application {

    private static Scene scene;
    private GeometryClient client;

    @Override
    public void start(Stage stage) throws IOException {
        // Initialise gRPC Client
        client = new GeometryClient("localhost", 50051);

        FXMLLoader fxmlLoader = new FXMLLoader(App.class.getResource("primary.fxml"));
        Parent root = fxmlLoader.load();

        // Inject client into controller
        PrimaryController controller = fxmlLoader.getController();
        controller.setClient(client);

        scene = new Scene(root, 800, 600);
        stage.setTitle("PolyNav - Delaunay Triangulation Client");
        stage.setScene(scene);
        stage.show();
    }

    @Override
    public void stop() throws Exception {
        if (client != null) {
            client.shutdown();
        }
        super.stop();
    }

    public static void main(String[] args) {
        launch();
    }
}
