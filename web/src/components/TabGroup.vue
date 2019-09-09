<script>
import TabContentGroup from 'components/TabContentGroup.vue';
import TabLinkGroup from 'components/TabLinkGroup.vue';

export default {
    functional: true,
    render: (createElement, context) => {
        const linkElements = context.children.filter((c) => c.tag && c.tag.includes('TabLink'));
        const contentElements = context.children.filter((c) => c.tag && c.tag.includes('TabContent'));
        return [
            createElement(TabLinkGroup, linkElements),
            createElement(TabContentGroup, contentElements),
        ];
    },
};
</script>
