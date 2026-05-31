<template>
  <view class="agreement-page">
    <view class="agreement-card">
      <view class="agreement-header">
        <text class="agreement-title">商家服务协议</text>
        <text class="agreement-desc">本协议已整合原隐私政策相关内容，请在使用平台服务前仔细阅读。</text>
      </view>

      <view
        v-for="section in agreementSections"
        :key="section.title"
        class="agreement-section"
      >
        <text class="section-title">{{ section.title }}</text>
        <text
          v-for="paragraph in section.paragraphs"
          :key="paragraph"
          class="section-paragraph"
        >
          {{ paragraph }}
        </text>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { APP_NAME } from '../../config/env'

interface AgreementSection {
  title: string
  paragraphs: string[]
}

const agreementSections: AgreementSection[] = [
  {
    title: '一、协议说明',
    paragraphs: [
      `1.1 本协议是您与${APP_NAME}平台之间关于使用商家管理服务所订立的协议，适用于商家账号登录、经营管理及相关服务使用行为。`,
      '1.2 您登录或继续使用平台服务，即视为已经阅读、理解并同意接受本协议全部内容。'
    ]
  },
  {
    title: '二、账号使用',
    paragraphs: [
      '2.1 商家账号由平台统一开通或分配，商家应妥善保管账号、密码及登录设备，避免账号泄露或被他人冒用。',
      '2.2 因商家自身保管不善导致的账号风险、数据泄露或经营损失，由商家自行承担相应责任。'
    ]
  },
  {
    title: '三、服务内容',
    paragraphs: [
      '3.1 平台为商家提供商品管理、订单管理、数据分析、配送管理、员工协作等经营辅助能力。',
      '3.2 平台有权根据业务发展需要调整、优化或变更服务内容，并通过页面公告、功能说明等方式通知商家。'
    ]
  },
  {
    title: '四、信息收集与使用',
    paragraphs: [
      '4.1 为向商家提供服务，平台会收集商家主动提供的信息，包括但不限于商家名称、联系方式、经营地址、账号信息及经营配置。',
      '4.2 平台会在必要范围内收集设备信息、操作日志、登录记录等信息，用于身份校验、服务安全保障、故障排查和体验优化。',
      '4.3 收集的信息将用于提供商家管理服务、处理订单交易、保障账号安全、改进产品功能及履行法律法规要求的义务。'
    ]
  },
  {
    title: '五、信息保护与共享',
    paragraphs: [
      '5.1 平台采用业界通行的安全技术与管理措施保护相关信息安全，并对内部接触信息的人员设置必要的权限控制与保密要求。',
      '5.2 未经商家同意，平台不会将相关信息用于本协议约定以外的用途；法律法规另有规定或监管机关要求的除外。',
      '5.3 为完成支付、配送、订单履约等必要服务，平台可能在必要范围内向支付机构、物流合作方等合作方共享订单相关信息。'
    ]
  },
  {
    title: '六、商家义务',
    paragraphs: [
      '6.1 商家应确保所售商品及经营活动符合国家法律法规要求，不得销售违禁、假冒伪劣商品，不得实施虚假宣传。',
      '6.2 商家应及时处理订单、履行售后义务，并保障消费者的合法权益。',
      '6.3 商家应保证提交给平台的信息真实、准确、完整，并在信息变更后及时更新。'
    ]
  },
  {
    title: '七、费用与结算',
    paragraphs: [
      '7.1 平台服务费用、结算规则及分账安排以平台与商家的实际开通配置及约定为准，平台按约定提供相关能力支持。',
      '7.2 商家应按照约定及时履行费用支付及结算配合义务。'
    ]
  },
  {
    title: '八、商家权利',
    paragraphs: [
      '8.1 商家有权访问、更正其提交的相关信息，并可在符合法律法规及平台规则的前提下申请撤回部分授权。',
      '8.2 如商家发现信息被违规使用、账号存在异常或平台服务存在安全风险，可及时联系平台处理。'
    ]
  },
  {
    title: '九、违约责任',
    paragraphs: [
      '9.1 任一方违反本协议约定，应依法承担相应责任；因商家违规经营、违法使用平台服务导致平台或第三方受损的，商家应承担赔偿责任。',
      '9.2 如商家存在严重违规、违法经营或危及平台安全的行为，平台有权暂停或终止相关服务，并保留进一步追责的权利。'
    ]
  },
  {
    title: '十、协议变更',
    paragraphs: [
      '10.1 平台有权根据业务需要或法律法规要求对本协议进行修订，修订内容可通过平台页面公示或其他合理方式通知商家。',
      '10.2 修订后的协议生效后，商家继续登录或使用平台服务的，视为接受更新后的协议内容。'
    ]
  }
]
</script>

<style>
.agreement-page {
  min-height: 100vh;
  padding: 24rpx;
  background: #f5f7fb;
}

.agreement-card {
  background: #ffffff;
  border-radius: 24rpx;
  padding: 32rpx;
  box-shadow: 0 8rpx 24rpx rgba(15, 23, 42, 0.05);
}

.agreement-header {
  margin-bottom: 24rpx;
}

.agreement-title {
  display: block;
  font-size: 40rpx;
  font-weight: 600;
  color: #1f2937;
  margin-bottom: 16rpx;
}

.agreement-desc {
  display: block;
  font-size: 26rpx;
  line-height: 1.8;
  color: #6b7280;
}

.agreement-section + .agreement-section {
  margin-top: 28rpx;
}

.section-title {
  display: block;
  font-size: 30rpx;
  font-weight: 600;
  color: #111827;
  margin-bottom: 12rpx;
}

.section-paragraph {
  display: block;
  font-size: 26rpx;
  line-height: 1.9;
  color: #4b5563;
}

.section-paragraph + .section-paragraph {
  margin-top: 10rpx;
}
</style>
