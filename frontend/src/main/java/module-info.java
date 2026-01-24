module fyp {
    requires transitive javafx.controls;
    requires javafx.fxml;
    requires protobuf.java;
    requires io.grpc;
    requires io.grpc.stub;
    requires io.grpc.protobuf;
    requires java.annotation;
    requires com.google.common;

    opens fyp to javafx.fxml;
    opens fyp.generated to protobuf.java;
    
    exports fyp;
    exports fyp.generated;
}
