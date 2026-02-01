module io.github.orbwarrior {
    requires transitive javafx.controls;
    requires javafx.fxml;
    requires com.google.protobuf;
    requires io.grpc;
    requires io.grpc.stub;
    requires io.grpc.protobuf;
    requires java.annotation;
    requires com.google.common;
    requires com.google.gson;

    opens io.github.orbwarrior to javafx.fxml, com.google.gson;
    opens fyp.generated to com.google.protobuf;
    
    exports io.github.orbwarrior;
    exports fyp.generated;
}