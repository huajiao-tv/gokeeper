
export function getErrorMessage(res){
    let errMessage = "";
    if (res.response) {
        if (res.response.data.error) {
            errMessage = res.response.data.error
        }else {
            errMessage = res.response.data
        }
    } else {
        errMessage = "请求失败"
    }
    return errMessage
}


export function successNotify(successMessage,border){
    border.$notify({
        title: "成功",
        message: successMessage,
        type: "success",
        duration: 2000
    })
}

export function errorNotify(errMessage,border){
    border.$notify({
        title: "失败",
        message: errMessage,
        type: "error",
        duration: 5000
    })
}
