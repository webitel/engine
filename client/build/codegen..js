const { codegen } = require('swagger-axios-codegen')

codegen({
    serviceNameSuffix: '',
    enumNamePrefix: 'Enum',
    methodNameMode: 'operationId',
    fileName: 'index.ts',
    useStaticMethod: true,
    useCustomerRequestInstance: false,

    source: require('../../script/api/api.swagger'),
    outputDir: '../codegen'
});