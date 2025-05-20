module fyp {
    requires javafx.controls;
    requires javafx.fxml;

    opens fyp to javafx.fxml;
    exports fyp;
}
