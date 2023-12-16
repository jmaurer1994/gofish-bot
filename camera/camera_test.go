package camera

import(
    "testing"
)

func TestCamera(t *testing.T){
    InitializeCamera("192.168.25.69", "admin", "")

    ResetLight()


    IncreaseLight()
    IncreaseLight()
    IncreaseLight()
    IncreaseLight()
    IncreaseLight()
    
    DecreaseLight()
    DecreaseLight()

    ResetLight()


    IncreaseLight()

}
