import cv2
from ultralytics import YOLO
import grpc
from concurrent import futures
import inferencetask_pb2
import inferencetask_pb2_grpc
import time


class TaskServiceServicer(inferencetask_pb2_grpc.TaskServiceServicer):

    def StreamResults(self, request, context):
        # Retrieve client information from context
        client_info = context.peer()
        print(f"[{client_info}] Client connected")
        frame_stats = []
        taskStart = time.time_ns()
        try:
            print(f"[{client_info}] Starting task:{request.task_name}")

            model = YOLO("")
            print("Loaded model")

            cap = cv2.VideoCapture(request.source)
            cap.set(cv2.CAP_PROP_BUFFERSIZE, 1)
            print(f"[{client_info}] Opened capture")
            while cap.isOpened():
                # Read a frame from the video
                frameStartTime = time.time_ns()

                success, frame = cap.read()
                if success:
                    # Run YOLOv8 tracking on the frame, persisting tracks between frames
                    results = model.track(
                        frame, persist=True, tracker="bytetrack.yaml")
                    result = self.GenerateTaskResultSet(results[0])
                    yield result
                    frameEndTime = time.time_ns()
                    frame_stats.append(frameEndTime - frameStartTime)

                else:
                    # Break the loop if the end of the video is reached
                    break
            cap.release()
        except Exception as e:
            print(f"[{client_info}] Error during tracking: {e}")
        finally:
            taskEnd = time.time_ns()
            print(f"[{client_info}] Tracking complete")
            print(f"[{client_info}] Processed {len(frame_stats)
                                               } frames in {(taskEnd - taskStart) / 1E9} s")
            print(f"[{client_info}] Average FPS: {
                  len(frame_stats) / ((taskEnd - taskStart) / 1E9)}")

            frameTime = 0
            for i, totalTime in enumerate(frame_stats):
                frameTime += totalTime
            print(f"[{client_info}] Average frame time: {
                  (frameTime / 1E6) / len(frame_stats)} ms")

    def GenerateTaskResultSet(self, result):
        summary_list = result.summary()
        if self.count < 5:
            print(summary_list)
            self.count += 1
        # Handle optional fields and populate gRPC TaskResult message
        task_result_set = inferencetask_pb2.TaskResultSet()

        for summary_dict in summary_list:
            task_result = inferencetask_pb2.TaskResult(
                # Use -1 as default for missing class
                class_id=summary_dict.get("class", -1),
                name=summary_dict.get("name", ""),
                track_id=summary_dict.get("track_id", -1),
                confidence=summary_dict.get("confidence", 0.0),
                speed=inferencetask_pb2.Speed(
                    inference=summary_dict.get(
                        "speed", {}).get("inference", 0.0),
                    postprocess=summary_dict.get(
                        "speed", {}).get("postprocess", 0.0),
                    preprocess=summary_dict.get(
                        "speed", {}).get("preprocess", 0.0)
                )
            )

            # Check if 'box' is present and populate if available
            box = summary_dict.get("box")
            if box:
                task_result.box.x1 = box.get("x1", 0.0)
                task_result.box.x2 = box.get("x2", 0.0)
                task_result.box.y1 = box.get("y1", 0.0)
                task_result.box.y2 = box.get("y2", 0.0)
                task_result.box.x3 = box.get("x3", 0.0)
                task_result.box.x4 = box.get("x4", 0.0)
                task_result.box.y3 = box.get("y3", 0.0)
                task_result.box.y4 = box.get("y4", 0.0)

            # Check if 'segments' is present and populate if available
            segments = summary_dict.get("segments")
            if segments:
                task_result.segments.x.extend(segments.get("x", []))
                task_result.segments.y.extend(segments.get("y", []))

            # Set the shape field as a repeated field
            task_result.shape.extend(summary_dict.get("shape", []))
            task_result_set.results.append(task_result)
        return task_result_set


def serve():
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    inferencetask_pb2_grpc.add_TaskServiceServicer_to_server(
        TaskServiceServicer(), server)
    server.add_insecure_port('[::]:51842')
    server.start()
    print("Server started, listening on port 51842.")
    server.wait_for_termination()


if __name__ == "__main__":
    serve()
