
export const lambdaHandler = async (event, context) => {
    console.log(event.body)

    try {
        return {
            'statusCode': 200,
            'body': event.body
        }
    } catch (err) {
        console.log(err)
        return err
    }
};
